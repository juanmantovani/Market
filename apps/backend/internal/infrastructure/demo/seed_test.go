package demo

import (
	"context"
	"errors"
	"testing"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/categories"
	"market/backend/internal/business/usecase/items"
	"market/backend/internal/infrastructure/repository/memory"
)

type repositories struct {
	categories *memory.CategoryRepository
	items      *memory.ItemRepository
}

func newCatalog() (Catalog, repositories) {
	store := memory.NewStore()
	repos := repositories{categories: memory.NewCategoryRepository(store), items: memory.NewItemRepository(store)}

	return Catalog{
		CreateCategory: categories.NewCreateCategory(repos.categories),
		CreateItem:     items.NewCreateItem(repos.items),
	}, repos
}

func TestSeedLoadsTheDemoCatalog(t *testing.T) {
	catalog, repos := newCatalog()

	if err := Seed(context.Background(), catalog); err != nil {
		t.Fatalf("Seed() error = %v", err)
	}

	categoryPage, err := categories.NewListCategories(repos.categories).Execute(context.Background(), categories.ListCategoriesInput{Actor: seeder})
	if err != nil || categoryPage.Total != len(demoCategories) {
		t.Fatalf("seeded %d categories (%v), want %d", categoryPage.Total, err, len(demoCategories))
	}

	itemPage, err := items.NewListItems(repos.items).Execute(context.Background(), items.ListItemsInput{Actor: seeder, Limit: 100})
	if err != nil || itemPage.Total != len(demoItems) {
		t.Fatalf("seeded %d items (%v), want %d", itemPage.Total, err, len(demoItems))
	}
	for _, item := range itemPage.Items {
		if item.CategoryID == "" {
			t.Fatalf("item %q has no category", item.Code)
		}
	}
}

func TestSeedReportsWhichRecordFailed(t *testing.T) {
	catalog, _ := newCatalog()
	if err := Seed(context.Background(), catalog); err != nil {
		t.Fatalf("first Seed() error = %v", err)
	}

	err := Seed(context.Background(), catalog)

	if !errors.Is(err, domain.ErrCategoryNameInUse) {
		t.Fatalf("second Seed() error = %v, want ErrCategoryNameInUse", err)
	}
}

func TestSeedReportsAFailedItem(t *testing.T) {
	catalog, _ := newCatalog()
	catalog.CreateItem = failingCreateItem{}

	err := Seed(context.Background(), catalog)

	if !errors.Is(err, domain.ErrStorageUnavailable) {
		t.Fatalf("Seed() error = %v, want ErrStorageUnavailable", err)
	}
}

type failingCreateItem struct{}

func (failingCreateItem) Execute(context.Context, items.CreateItemInput) (domain.Item, error) {
	return domain.Item{}, domain.ErrStorageUnavailable
}
