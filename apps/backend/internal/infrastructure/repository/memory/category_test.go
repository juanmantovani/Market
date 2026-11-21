package memory

import (
	"context"
	"errors"
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
		Sellers:    NewSellerRepository(store),
		Stock:      NewStockRepository(store),
		Sales:      NewSaleRepository(store),
		Tx:         store,
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

func TestSellerRepositoryContract(t *testing.T) {
	repositorytest.SellerRepository(t, newRepositories)
}

func TestSaleRepositoryContract(t *testing.T) {
	repositorytest.SaleRepository(t, newRepositories)
}

func TestStockRepositoryContract(t *testing.T) {
	repositorytest.StockRepository(t, newRepositories)
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
