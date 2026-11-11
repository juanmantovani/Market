package memory

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"market/backend/internal/infrastructure/repository/repositorytest"
)

func newRepositories(*testing.T) repositorytest.Repositories {
	store := NewStore()
	return repositorytest.Repositories{
		Categories: NewCategoryRepository(store),
		Items:      NewItemRepository(store),
		Customers:  NewCustomerRepository(store),
		Suppliers:  NewSupplierRepository(store),
	}
}

func TestCategoryRepositoryContract(t *testing.T) {
	repositorytest.CategoryRepository(t, newRepositories)
}

func TestItemRepositoryContract(t *testing.T) {
	repositorytest.ItemRepository(t, newRepositories)
}

func TestCustomerRepositoryContract(t *testing.T) {
	repositorytest.CustomerRepository(t, newRepositories)
}

func TestSupplierRepositoryContract(t *testing.T) {
	repositorytest.SupplierRepository(t, newRepositories)
}

func TestRepositoriesHonorCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repositories := newRepositories(t)

	if _, err := repositories.Categories.Get(ctx, "any"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Categories.Get() error = %v, want context.Canceled", err)
	}
	if _, err := repositories.Items.Get(ctx, "any"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Items.Get() error = %v, want context.Canceled", err)
	}
}

func TestNewIDIsAUUIDv7(t *testing.T) {
	uuidV7 := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	first, second := newID(), newID()

	if !uuidV7.MatchString(first) {
		t.Fatalf("newID() = %q, want a UUIDv7", first)
	}
	if first == second {
		t.Fatal("newID() repeated an ID")
	}
}
