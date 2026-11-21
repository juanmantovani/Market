package dependencies

import (
	"context"
	"testing"

	"market/backend/internal/infrastructure/environments"
)

func TestNewBuildsTheMemoryGraph(t *testing.T) {
	container, err := New(context.Background(), environments.Server{StorageDriver: environments.StorageMemory})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	handlers := container.Handlers()
	if handlers.Ready == nil ||
		handlers.ListCategories == nil || handlers.CreateCategory == nil || handlers.DeleteCategory == nil ||
		handlers.ListStockLevels == nil || handlers.CreateStockEntry == nil || handlers.CreateStockAdjustment == nil || handlers.CreateStockInitial == nil ||
		handlers.CreateSale == nil || handlers.ListSales == nil || handlers.GetSale == nil || handlers.VoidSale == nil || handlers.GetDashboardSummary == nil {
		t.Fatalf("Handlers() left routes unwired: %+v", handlers)
	}
	if container.RegisterInitialStock == nil {
		t.Fatal("New() left RegisterInitialStock unwired")
	}
	if err := container.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() error = %v, want the memory driver to always be ready", err)
	}
}

// pgxpool.New only parses the connection string; it doesn't connect, so this
// builds the whole postgres graph without a real database.
func TestNewBuildsThePostgresGraph(t *testing.T) {
	container, err := New(context.Background(), environments.Server{
		StorageDriver: environments.StoragePostgres,
		DatabaseURL:   "postgres://user:pass@localhost:5432/market",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	handlers := container.Handlers()
	if handlers.Ready == nil || handlers.ListCustomers == nil || handlers.CreateCustomer == nil {
		t.Fatalf("Handlers() left the customer routes unwired: %+v", handlers)
	}
	if container.Ping == nil {
		t.Fatal("New() left Ping unwired")
	}
}

func TestNewRejectsAnUnknownDriver(t *testing.T) {
	if _, err := New(context.Background(), environments.Server{StorageDriver: "sqlite"}); err == nil {
		t.Fatal("New() error = nil, want an unsupported driver error")
	}
}
