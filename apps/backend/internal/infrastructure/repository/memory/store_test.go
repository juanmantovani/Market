package memory

import (
	"context"
	"errors"
	"sync"
	"testing"
)

func TestWithinTxKeepsWritesOnSuccess(t *testing.T) {
	store := NewStore()
	table := map[string]int{"a": 1}

	err := store.WithinTx(context.Background(), func(ctx context.Context) error {
		return store.run(ctx, func(current *tx) error {
			put(current, table, "a", 2)
			put(current, table, "b", 3)
			remove(current, table, "missing")
			return nil
		})
	})
	if err != nil {
		t.Fatalf("WithinTx() error = %v", err)
	}

	if table["a"] != 2 || table["b"] != 3 || len(table) != 2 {
		t.Fatalf("table = %v, want map[a:2 b:3]", table)
	}
}

func TestWithinTxRollsBackOnError(t *testing.T) {
	store := NewStore()
	table := map[string]int{"a": 1, "z": 26}
	failure := errors.New("insufficient stock")

	err := store.WithinTx(context.Background(), func(ctx context.Context) error {
		if err := store.run(ctx, func(current *tx) error {
			put(current, table, "a", 2)
			put(current, table, "a", 3)
			put(current, table, "b", 4)
			remove(current, table, "z")
			return nil
		}); err != nil {
			return err
		}
		return failure
	})
	if !errors.Is(err, failure) {
		t.Fatalf("WithinTx() error = %v, want %v", err, failure)
	}

	want := map[string]int{"a": 1, "z": 26}
	if len(table) != len(want) || table["a"] != 1 || table["z"] != 26 {
		t.Fatalf("table after rollback = %v, want %v", table, want)
	}
}

func TestWithinTxRollsBackAndRepanics(t *testing.T) {
	store := NewStore()
	table := map[string]int{}

	defer func() {
		if recover() == nil {
			t.Fatal("WithinTx swallowed the panic")
		}
		if len(table) != 0 {
			t.Fatalf("table after panic = %v, want empty", table)
		}
		if err := store.WithinTx(context.Background(), func(context.Context) error { return nil }); err != nil {
			t.Fatalf("store unusable after panic: %v", err)
		}
	}()

	_ = store.WithinTx(context.Background(), func(ctx context.Context) error {
		_ = store.run(ctx, func(current *tx) error {
			put(current, table, "a", 1)
			return nil
		})
		panic("boom")
	})
}

func TestNestedWithinTxJoinsTheOuterTransaction(t *testing.T) {
	store := NewStore()
	table := map[string]int{}
	failure := errors.New("outer failure")

	err := store.WithinTx(context.Background(), func(ctx context.Context) error {
		if err := store.WithinTx(ctx, func(ctx context.Context) error {
			return store.run(ctx, func(current *tx) error {
				put(current, table, "inner", 1)
				return nil
			})
		}); err != nil {
			return err
		}
		return failure
	})
	if !errors.Is(err, failure) {
		t.Fatalf("WithinTx() error = %v, want %v", err, failure)
	}
	if len(table) != 0 {
		t.Fatalf("inner write survived the outer rollback: %v", table)
	}
}

func TestTransactionOfAnotherStoreIsNotJoined(t *testing.T) {
	first, second := NewStore(), NewStore()

	err := first.WithinTx(context.Background(), func(ctx context.Context) error {
		if second.txFrom(ctx) != nil {
			t.Fatal("second store joined the first store's transaction")
		}
		return second.run(ctx, func(*tx) error { return nil })
	})
	if err != nil {
		t.Fatalf("WithinTx() error = %v", err)
	}
}

func TestWithinTxRejectsACancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	called := false
	err := NewStore().WithinTx(ctx, func(context.Context) error {
		called = true
		return nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("WithinTx() error = %v, want context.Canceled", err)
	}
	if called {
		t.Fatal("fn ran with a cancelled context")
	}
}

func TestWithinTxSerializesConcurrentWriters(t *testing.T) {
	store := NewStore()
	table := map[string]int{"stock": 0}

	const writers = 50
	var wg sync.WaitGroup
	for range writers {
		wg.Go(func() {
			_ = store.run(context.Background(), func(current *tx) error {
				put(current, table, "stock", table["stock"]+1)
				return nil
			})
		})
	}
	wg.Wait()

	if table["stock"] != writers {
		t.Fatalf("stock = %d, want %d", table["stock"], writers)
	}
}
