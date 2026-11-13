package demo

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/categories"
	"market/backend/internal/business/usecase/customers"
	"market/backend/internal/business/usecase/items"
	"market/backend/internal/business/usecase/sales"
	"market/backend/internal/business/usecase/sellers"
	"market/backend/internal/business/usecase/stock"
	"market/backend/internal/business/usecase/suppliers"
	"market/backend/internal/infrastructure/repository/memory"
)

type repositories struct {
	categories *memory.CategoryRepository
	items      *memory.ItemRepository
	customers  *memory.CustomerRepository
	suppliers  *memory.SupplierRepository
	sellers    *memory.SellerRepository
	stock      *memory.StockRepository
	sales      *memory.SaleRepository
}

func newUseCases() (UseCases, repositories) {
	store := memory.NewStore()
	repos := repositories{
		categories: memory.NewCategoryRepository(store),
		items:      memory.NewItemRepository(store),
		customers:  memory.NewCustomerRepository(store),
		suppliers:  memory.NewSupplierRepository(store),
		sellers:    memory.NewSellerRepository(store),
		stock:      memory.NewStockRepository(store),
		sales:      memory.NewSaleRepository(store),
	}

	return UseCases{
		CreateCategory: categories.NewCreateCategory(repos.categories),
		CreateItem:     items.NewCreateItem(repos.items),
		CreateCustomer: customers.NewCreateCustomer(repos.customers),
		CreateSupplier: suppliers.NewCreateSupplier(repos.suppliers),
		CreateSeller:   sellers.NewCreateSeller(repos.sellers),

		RegisterInitialStock:    stock.NewRegisterInitialStock(store, repos.stock, repos.items),
		RegisterStockEntry:      stock.NewRegisterStockEntry(store, repos.stock, repos.items, repos.suppliers),
		RegisterStockAdjustment: stock.NewRegisterStockAdjustment(store, repos.stock, repos.items),

		CreateSale: sales.NewCreateSale(store, repos.sales, repos.items, repos.customers, repos.sellers, repos.stock, domain.StockPolicy{}),
		VoidSale:   sales.NewVoidSale(store, repos.sales, repos.stock),
	}, repos
}

func TestSeedLoadsTheDemoCatalog(t *testing.T) {
	useCases, repos := newUseCases()

	if err := Seed(context.Background(), useCases); err != nil {
		t.Fatalf("Seed() error = %v", err)
	}

	customerPage, err := customers.NewListCustomers(repos.customers).Execute(context.Background(), customers.ListCustomersInput{Actor: seeder})
	if err != nil || customerPage.Total != len(demoCustomers)+1 {
		t.Fatalf("seeded %d customers (%v), want %d plus Consumidor Final", customerPage.Total, err, len(demoCustomers))
	}

	supplierPage, err := suppliers.NewListSuppliers(repos.suppliers).Execute(context.Background(), suppliers.ListSuppliersInput{Actor: seeder})
	if err != nil || supplierPage.Total != len(demoSuppliers) {
		t.Fatalf("seeded %d suppliers (%v), want %d", supplierPage.Total, err, len(demoSuppliers))
	}

	sellerPage, err := sellers.NewListSellers(repos.sellers).Execute(context.Background(), sellers.ListSellersInput{Actor: seeder})
	if err != nil || sellerPage.Total != len(demoSellers) {
		t.Fatalf("seeded %d sellers (%v), want %d", sellerPage.Total, err, len(demoSellers))
	}

	categoryPage, err := categories.NewListCategories(repos.categories).Execute(context.Background(), categories.ListCategoriesInput{Actor: seeder})
	if err != nil || categoryPage.Total != len(demoCategories) {
		t.Fatalf("seeded %d categories (%v), want %d", categoryPage.Total, err, len(demoCategories))
	}

	itemPage, err := items.NewListItems(repos.items).Execute(context.Background(), items.ListItemsInput{Actor: seeder, Limit: 100})
	if err != nil || itemPage.Total != len(demoItems) {
		t.Fatalf("seeded %d items (%v), want %d", itemPage.Total, err, len(demoItems))
	}
	suppliersByCode := map[string]int{}
	for _, item := range itemPage.Items {
		if item.CategoryID == "" {
			t.Fatalf("item %q has no category", item.Code)
		}
		suppliersByCode[item.Code] = len(item.SupplierIDs)
	}
	if suppliersByCode["ALM-001"] != 2 || suppliersByCode["LAC-001"] != 1 || suppliersByCode["FIA-001"] != 0 {
		t.Fatalf("suppliers per item = %v", suppliersByCode)
	}
}

func TestSeedLoadsTheDemoStock(t *testing.T) {
	useCases, repos := newUseCases()
	if err := Seed(context.Background(), useCases); err != nil {
		t.Fatalf("Seed() error = %v", err)
	}

	levels, err := stock.NewListStockLevels(repos.stock).Execute(context.Background(), stock.ListStockLevelsInput{Actor: seeder, Limit: 100})
	if err != nil || levels.Total != len(demoItems) {
		t.Fatalf("stock levels = %d (%v), want one per item", levels.Total, err)
	}
	quantities := map[string]domain.Quantity{}
	var low []string
	for _, level := range levels.Items {
		quantities[level.Item.Code] = level.Level.Quantity
		if level.IsLow() {
			low = append(low, level.Item.Code)
		}
	}
	if quantities["BEB-001"] != domain.WholeQuantity(28) || quantities["KIO-001"] != domain.WholeQuantity(57) || quantities["FIA-001"] != 1_550 {
		t.Fatalf("stock per item = %v", quantities)
	}
	if fmt.Sprint(low) != "[LAC-001 FIA-001 ALM-001]" {
		t.Fatalf("items below their minimum = %v", low)
	}

	saleLines, voidedLines := 0, 0
	for _, sale := range demoSales {
		saleLines += len(sale.lines)
		if sale.voidReason != "" {
			voidedLines += len(sale.lines)
		}
	}
	movements, err := stock.NewListStockMovements(repos.stock).Execute(context.Background(), stock.ListStockMovementsInput{Actor: seeder, Limit: 100})
	if err != nil || movements.Total != len(demoItems)+2+saleLines+voidedLines {
		t.Fatalf("movements = %d (%v), want an opening one per item, an entry, an adjustment one per sale line and one per voided line", movements.Total, err)
	}
	byType := map[domain.MovementType]int{}
	for _, movement := range movements.Items {
		byType[movement.Type]++
	}
	if byType[domain.MovementEntry] != 1 || byType[domain.MovementAdjustment] != 1 || byType[domain.MovementSale] != saleLines || byType[domain.MovementSaleVoid] != voidedLines || voidedLines == 0 {
		t.Fatalf("movements by type = %v", byType)
	}
}

func TestSeedLoadsTheDemoSales(t *testing.T) {
	useCases, repos := newUseCases()
	if err := Seed(context.Background(), useCases); err != nil {
		t.Fatalf("Seed() error = %v", err)
	}

	for i, want := range demoSales {
		sale, err := repos.sales.GetByIdempotencyKey(context.Background(), fmt.Sprintf("demo-sale-%d", i+1))
		if err != nil {
			t.Fatalf("demo sale %d: %v", i+1, err)
		}
		if sale.Number != int64(i+1) || sale.Total != want.total || len(sale.Lines) != len(want.lines) || (want.seller == "") != (sale.Seller == nil) {
			t.Fatalf("demo sale %d = %+v", i+1, sale)
		}
		if want.customer == "" && sale.Customer.ID != domain.FinalConsumerID {
			t.Fatalf("demo sale %d without customer went to %+v", i+1, sale.Customer)
		}
		wantStatus := domain.SaleCompleted
		if want.voidReason != "" {
			wantStatus = domain.SaleVoided
		}
		if sale.Status != wantStatus {
			t.Fatalf("demo sale %d status = %s", i+1, sale.Status)
		}
	}
}

func TestSeedReportsAFailedSale(t *testing.T) {
	useCases, _ := newUseCases()
	useCases.CreateSale = failingCreateSale{}

	err := Seed(context.Background(), useCases)

	if !errors.Is(err, domain.ErrStorageUnavailable) {
		t.Fatalf("Seed() error = %v, want ErrStorageUnavailable", err)
	}
}

type failingCreateSale struct{}

func (failingCreateSale) Execute(context.Context, sales.CreateSaleInput) (domain.Sale, error) {
	return domain.Sale{}, domain.ErrStorageUnavailable
}

func TestSeedReportsAFailedStockStep(t *testing.T) {
	tests := []struct {
		name    string
		arrange func(*UseCases)
	}{
		{"initial stock", func(useCases *UseCases) { useCases.RegisterInitialStock = failingInitialStock{} }},
		{"entry", func(useCases *UseCases) { useCases.RegisterStockEntry = failingStockEntry{} }},
		{"adjustment", func(useCases *UseCases) { useCases.RegisterStockAdjustment = failingStockAdjustment{} }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useCases, _ := newUseCases()
			test.arrange(&useCases)

			err := Seed(context.Background(), useCases)

			if !errors.Is(err, domain.ErrStorageUnavailable) {
				t.Fatalf("Seed() error = %v, want ErrStorageUnavailable", err)
			}
		})
	}
}

type failingInitialStock struct{}

func (failingInitialStock) Execute(context.Context, stock.RegisterInitialStockInput) (domain.StockMovement, error) {
	return domain.StockMovement{}, domain.ErrStorageUnavailable
}

type failingStockEntry struct{}

func (failingStockEntry) Execute(context.Context, stock.RegisterStockEntryInput) ([]domain.StockMovement, error) {
	return nil, domain.ErrStorageUnavailable
}

type failingStockAdjustment struct{}

func (failingStockAdjustment) Execute(context.Context, stock.RegisterStockAdjustmentInput) (domain.StockMovement, error) {
	return domain.StockMovement{}, domain.ErrStorageUnavailable
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

func TestSeedReportsAFailedSeller(t *testing.T) {
	useCases, _ := newUseCases()
	useCases.CreateSeller = failingCreateSeller{}

	err := Seed(context.Background(), useCases)

	if !errors.Is(err, domain.ErrStorageUnavailable) {
		t.Fatalf("Seed() error = %v, want ErrStorageUnavailable", err)
	}
}

type failingCreateSeller struct{}

func (failingCreateSeller) Execute(context.Context, sellers.CreateSellerInput) (domain.Seller, error) {
	return domain.Seller{}, domain.ErrStorageUnavailable
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

func TestSeedReportsAFailedVoid(t *testing.T) {
	useCases, _ := newUseCases()
	useCases.VoidSale = failingVoidSale{}

	err := Seed(context.Background(), useCases)

	if !errors.Is(err, domain.ErrStorageUnavailable) {
		t.Fatalf("Seed() error = %v, want ErrStorageUnavailable", err)
	}
}

type failingVoidSale struct{}

func (failingVoidSale) Execute(context.Context, sales.VoidSaleInput) (domain.Sale, error) {
	return domain.Sale{}, domain.ErrStorageUnavailable
}
