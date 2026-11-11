package demo

import (
	"context"
	"errors"
	"testing"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/categories"
	"market/backend/internal/business/usecase/customers"
	"market/backend/internal/business/usecase/items"
	"market/backend/internal/business/usecase/suppliers"
	"market/backend/internal/infrastructure/repository/memory"
)

type repositories struct {
	categories *memory.CategoryRepository
	items      *memory.ItemRepository
	customers  *memory.CustomerRepository
	suppliers  *memory.SupplierRepository
}

func newUseCases() (UseCases, repositories) {
	store := memory.NewStore()
	repos := repositories{
		categories: memory.NewCategoryRepository(store),
		items:      memory.NewItemRepository(store),
		customers:  memory.NewCustomerRepository(store),
		suppliers:  memory.NewSupplierRepository(store),
	}

	return UseCases{
		CreateCategory: categories.NewCreateCategory(repos.categories),
		CreateItem:     items.NewCreateItem(repos.items),
		CreateCustomer: customers.NewCreateCustomer(repos.customers),
		CreateSupplier: suppliers.NewCreateSupplier(repos.suppliers),
	}, repos
}

func TestSeedLoadsTheDemoCatalog(t *testing.T) {
	useCases, repos := newUseCases()

	if err := Seed(context.Background(), useCases); err != nil {
		t.Fatalf("Seed() error = %v", err)
	}

	customerPage, err := customers.NewListCustomers(repos.customers).Execute(context.Background(), customers.ListCustomersInput{Actor: seeder})
	if err != nil || customerPage.Total != len(demoCustomers) {
		t.Fatalf("seeded %d customers (%v), want %d", customerPage.Total, err, len(demoCustomers))
	}

	supplierPage, err := suppliers.NewListSuppliers(repos.suppliers).Execute(context.Background(), suppliers.ListSuppliersInput{Actor: seeder})
	if err != nil || supplierPage.Total != len(demoSuppliers) {
		t.Fatalf("seeded %d suppliers (%v), want %d", supplierPage.Total, err, len(demoSuppliers))
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
	useCases, _ := newUseCases()
	if err := Seed(context.Background(), useCases); err != nil {
		t.Fatalf("first Seed() error = %v", err)
	}

	err := Seed(context.Background(), useCases)

	if !errors.Is(err, domain.ErrCategoryNameInUse) {
		t.Fatalf("second Seed() error = %v, want ErrCategoryNameInUse", err)
	}
}

func TestSeedReportsAFailedItem(t *testing.T) {
	useCases, _ := newUseCases()
	useCases.CreateItem = failingCreateItem{}

	err := Seed(context.Background(), useCases)

	if !errors.Is(err, domain.ErrStorageUnavailable) {
		t.Fatalf("Seed() error = %v, want ErrStorageUnavailable", err)
	}
}

func TestSeedReportsAFailedCustomer(t *testing.T) {
	useCases, _ := newUseCases()
	useCases.CreateCustomer = failingCreateCustomer{}

	err := Seed(context.Background(), useCases)

	if !errors.Is(err, domain.ErrStorageUnavailable) {
		t.Fatalf("Seed() error = %v, want ErrStorageUnavailable", err)
	}
}

func TestSeedReportsAFailedSupplier(t *testing.T) {
	useCases, _ := newUseCases()
	useCases.CreateSupplier = failingCreateSupplier{}

	err := Seed(context.Background(), useCases)

	if !errors.Is(err, domain.ErrStorageUnavailable) {
		t.Fatalf("Seed() error = %v, want ErrStorageUnavailable", err)
	}
}

type failingCreateSupplier struct{}

func (failingCreateSupplier) Execute(context.Context, suppliers.CreateSupplierInput) (domain.Supplier, error) {
	return domain.Supplier{}, domain.ErrStorageUnavailable
}

type failingCreateCustomer struct{}

func (failingCreateCustomer) Execute(context.Context, customers.CreateCustomerInput) (domain.Customer, error) {
	return domain.Customer{}, domain.ErrStorageUnavailable
}

type failingCreateItem struct{}

func (failingCreateItem) Execute(context.Context, items.CreateItemInput) (domain.Item, error) {
	return domain.Item{}, domain.ErrStorageUnavailable
}
