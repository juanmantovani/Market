package postgres

import (
	"context"
	"errors"
	"testing"

	"market/backend/internal/infrastructure/repository/repositorytest"
)

func TestCustomerRepositoryContract(t *testing.T) {
	requireTestDatabase(t)
	repositorytest.CustomerRepository(t, newRepositories)
}

func TestCustomerRepositoryHonorsCancellation(t *testing.T) {
	requireTestDatabase(t)
	repositories := newRepositories(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := repositories.Customers.Get(ctx, "any"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Get() error = %v, want context.Canceled", err)
	}
}
