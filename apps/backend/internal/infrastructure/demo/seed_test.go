package demo

import (
	"context"
	"errors"
	"testing"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/categories"
	"market/backend/internal/infrastructure/repository/memory"
)

func TestSeedLoadsTheDemoCatalog(t *testing.T) {
	repository := memory.NewCategoryRepository(memory.NewStore())
	catalog := Catalog{CreateCategory: categories.NewCreateCategory(repository)}

	if err := Seed(context.Background(), catalog); err != nil {
		t.Fatalf("Seed() error = %v", err)
	}

	list := categories.NewListCategories(repository)
	page, err := list.Execute(context.Background(), categories.ListCategoriesInput{Actor: seeder})
	if err != nil {
		t.Fatalf("listing seeded categories: %v", err)
	}
	if page.Total != len(demoCategories) {
		t.Fatalf("seeded %d categories, want %d", page.Total, len(demoCategories))
	}
}

func TestSeedReportsWhichRecordFailed(t *testing.T) {
	repository := memory.NewCategoryRepository(memory.NewStore())
	catalog := Catalog{CreateCategory: categories.NewCreateCategory(repository)}
	if err := Seed(context.Background(), catalog); err != nil {
		t.Fatalf("first Seed() error = %v", err)
	}

	err := Seed(context.Background(), catalog)

	if !errors.Is(err, domain.ErrCategoryNameInUse) {
		t.Fatalf("second Seed() error = %v, want ErrCategoryNameInUse", err)
	}
}
