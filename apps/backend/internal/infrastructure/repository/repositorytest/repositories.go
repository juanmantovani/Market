package repositorytest

import (
	"testing"

	"market/backend/internal/business/gateway"
)

// Repositories are the repositories of one adapter sharing the same
// storage, so a contract can check rules that span entities (an item needs
// its category; a category with items can't be deleted).
type Repositories struct {
	Categories gateway.CategoryRepository
	Items      gateway.ItemRepository
	Customers  gateway.CustomerRepository
	Suppliers  gateway.SupplierRepository
	Sellers    gateway.SellerRepository
	Stock      gateway.StockRepository
	// Tx is the adapter's unit of work over these same repositories.
	Tx gateway.TxManager
}

// NewRepositories must return empty repositories each time it's called.
type NewRepositories func(t *testing.T) Repositories
