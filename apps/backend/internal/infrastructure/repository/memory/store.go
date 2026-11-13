package memory

import (
	"context"
	"sync"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

var _ gateway.TxManager = (*Store)(nil)

// Store holds every in-memory table behind a single lock, so a transaction
// is serializable: WithinTx keeps the lock for the whole unit of work.
// Writes register an undo step; if the transaction fails, the steps run in
// reverse and the tables end up exactly as they were.
//
// Tables keep values, and repositories copy any pointer or slice they store
// or return, so callers can't mutate a table behind the lock.
type Store struct {
	mu sync.Mutex

	categories map[string]domain.Category
	items      map[string]domain.Item
	customers  map[string]domain.Customer
	suppliers  map[string]domain.Supplier
	sellers    map[string]domain.Seller

	stockLevels map[string]domain.StockLevel
	// stockMovements only grows: movements are never edited or deleted, and
	// their order is the order they were recorded in.
	stockMovements []domain.StockMovement

	sales map[string]domain.Sale
	// lastSaleNumber is what a sequence is in PostgreSQL. Unlike a
	// sequence, a failed unit of work gives its number back.
	lastSaleNumber int64
}

// NewStore returns an empty store but for the system customer, which every
// storage starts with (in PostgreSQL, a migration inserts it).
func NewStore() *Store {
	finalConsumer := domain.FinalConsumer(time.Now().UTC())

	return &Store{
		categories:  map[string]domain.Category{},
		items:       map[string]domain.Item{},
		customers:   map[string]domain.Customer{finalConsumer.ID: finalConsumer},
		suppliers:   map[string]domain.Supplier{},
		sellers:     map[string]domain.Seller{},
		stockLevels: map[string]domain.StockLevel{},
		sales:       map[string]domain.Sale{},
	}
}

type txKey struct{}

type tx struct {
	store *Store
	undo  []func()
}

// WithinTx runs fn holding the store lock. A ctx already inside one of this
// store's transactions joins it instead of deadlocking on the lock.
func (store *Store) WithinTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	if store.txFrom(ctx) != nil {
		return fn(ctx)
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	current := &tx{store: store}
	defer func() {
		if recovered := recover(); recovered != nil {
			current.rollback()
			panic(recovered)
		}
		if err != nil {
			current.rollback()
		}
	}()

	return fn(context.WithValue(ctx, txKey{}, current))
}

// run executes a repository operation under the store lock, joining the
// caller's transaction when there is one.
func (store *Store) run(ctx context.Context, operation func(current *tx) error) error {
	if current := store.txFrom(ctx); current != nil {
		return operation(current)
	}

	return store.WithinTx(ctx, func(ctx context.Context) error {
		return operation(store.txFrom(ctx))
	})
}

func (store *Store) txFrom(ctx context.Context) *tx {
	current, ok := ctx.Value(txKey{}).(*tx)
	if !ok || current.store != store {
		return nil
	}

	return current
}

func (current *tx) rollback() {
	for i := len(current.undo) - 1; i >= 0; i-- {
		current.undo[i]()
	}
	current.undo = nil
}

// put writes value under key and records how to undo it.
func put[K comparable, V any](current *tx, table map[K]V, key K, value V) {
	previous, existed := table[key]
	table[key] = value
	current.undo = append(current.undo, func() {
		if existed {
			table[key] = previous
		} else {
			delete(table, key)
		}
	})
}

// set writes value to a single-value cell, such as a counter, and records
// how to undo it.
func set[V any](current *tx, cell *V, value V) {
	previous := *cell
	*cell = value
	current.undo = append(current.undo, func() {
		*cell = previous
	})
}

// appendRow appends value to an append-only table and records how to undo
// it. Undo steps run in reverse, so dropping the last row is always right.
func appendRow[V any](current *tx, table *[]V, value V) {
	*table = append(*table, value)
	current.undo = append(current.undo, func() {
		*table = (*table)[:len(*table)-1]
	})
}

// remove deletes key and records how to undo it.
func remove[K comparable, V any](current *tx, table map[K]V, key K) {
	previous, existed := table[key]
	if !existed {
		return
	}
	delete(table, key)
	current.undo = append(current.undo, func() {
		table[key] = previous
	})
}
