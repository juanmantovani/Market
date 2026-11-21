package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"market/backend/internal/business/domain"
)

func TestWithinTxCommitsOnSuccess(t *testing.T) {
	requireTestDatabase(t)
	repositories := newRepositories(t)
	manager := repositories.Tx

	var created string
	err := manager.WithinTx(context.Background(), func(ctx context.Context) error {
		customer, err := repositories.Customers.Create(ctx, newTestCustomer(t, "Ana"))
		created = customer.ID
		return err
	})
	if err != nil {
		t.Fatalf("WithinTx() error = %v", err)
	}

	if _, err := repositories.Customers.Get(context.Background(), created); err != nil {
		t.Fatalf("Get() after commit error = %v", err)
	}
}

func TestWithinTxRollsBackOnError(t *testing.T) {
	requireTestDatabase(t)
	repositories := newRepositories(t)
	manager := repositories.Tx
	failure := errors.New("insufficient stock")

	var created string
	err := manager.WithinTx(context.Background(), func(ctx context.Context) error {
		customer, err := repositories.Customers.Create(ctx, newTestCustomer(t, "Ana"))
		if err != nil {
			return err
		}
		created = customer.ID
		return failure
	})
	if !errors.Is(err, failure) {
		t.Fatalf("WithinTx() error = %v, want %v", err, failure)
	}

	if _, err := repositories.Customers.Get(context.Background(), created); err == nil {
		t.Fatal("Get() after rollback found the customer")
	}
}

func TestNestedWithinTxJoinsTheOuterTransaction(t *testing.T) {
	requireTestDatabase(t)
	repositories := newRepositories(t)
	manager := repositories.Tx
	failure := errors.New("outer failure")

	var created string
	err := manager.WithinTx(context.Background(), func(ctx context.Context) error {
		if err := manager.WithinTx(ctx, func(ctx context.Context) error {
			customer, err := repositories.Customers.Create(ctx, newTestCustomer(t, "Ana"))
			created = customer.ID
			return err
		}); err != nil {
			return err
		}
		return failure
	})
	if !errors.Is(err, failure) {
		t.Fatalf("WithinTx() error = %v, want %v", err, failure)
	}

	if _, err := repositories.Customers.Get(context.Background(), created); err == nil {
		t.Fatal("Get() after the outer rollback found the customer from the inner transaction")
	}
}

func newTestCustomer(t *testing.T, name string) domain.Customer {
	t.Helper()
	customer, err := domain.NewCustomer(domain.CustomerFields{
		Name:         name,
		DocumentType: domain.DocumentNone,
		TaxCondition: domain.TaxFinalConsumer,
	}, time.Now().UTC())
	if err != nil {
		t.Fatalf("domain.NewCustomer(%q) error = %v", name, err)
	}
	return customer
}
