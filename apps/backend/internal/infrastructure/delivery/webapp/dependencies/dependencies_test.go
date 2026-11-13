package dependencies

import (
	"testing"

	"market/backend/internal/infrastructure/environments"
)

func TestNewBuildsTheMemoryGraph(t *testing.T) {
	container, err := New(environments.Server{StorageDriver: environments.StorageMemory})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	handlers := container.Handlers()
	if handlers.ListCategories == nil || handlers.CreateCategory == nil || handlers.DeleteCategory == nil ||
		handlers.ListStockLevels == nil || handlers.CreateStockEntry == nil || handlers.CreateStockAdjustment == nil ||
		handlers.CreateSale == nil || handlers.ListSales == nil || handlers.GetSale == nil || handlers.VoidSale == nil {
		t.Fatalf("Handlers() left routes unwired: %+v", handlers)
	}
	if container.RegisterInitialStock == nil {
		t.Fatal("New() left RegisterInitialStock unwired")
	}
}

func TestNewRejectsAnUnknownDriver(t *testing.T) {
	if _, err := New(environments.Server{StorageDriver: "sqlite"}); err == nil {
		t.Fatal("New() error = nil, want an unsupported driver error")
	}
}
