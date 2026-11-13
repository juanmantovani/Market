package dependencies

import (
	"fmt"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
	"market/backend/internal/business/usecase/categories"
	"market/backend/internal/business/usecase/customers"
	"market/backend/internal/business/usecase/items"
	"market/backend/internal/business/usecase/sales"
	"market/backend/internal/business/usecase/sellers"
	"market/backend/internal/business/usecase/stock"
	"market/backend/internal/business/usecase/suppliers"
	"market/backend/internal/infrastructure/delivery/webapp/handler"
	"market/backend/internal/infrastructure/delivery/webapp/route"
	"market/backend/internal/infrastructure/environments"
	"market/backend/internal/infrastructure/repository/memory"
)

// Container is the object graph: repositories → use cases → handlers. It
// only builds objects; reading configuration, registering routes and running
// the server belong to internal/app.
type Container struct {
	CreateCategory categories.CreateCategory
	ListCategories categories.ListCategories
	GetCategory    categories.GetCategory
	UpdateCategory categories.UpdateCategory
	DeleteCategory categories.DeleteCategory

	CreateItem items.CreateItem
	ListItems  items.ListItems
	GetItem    items.GetItem
	UpdateItem items.UpdateItem
	DeleteItem items.DeleteItem

	CreateCustomer customers.CreateCustomer
	ListCustomers  customers.ListCustomers
	GetCustomer    customers.GetCustomer
	UpdateCustomer customers.UpdateCustomer
	DeleteCustomer customers.DeleteCustomer

	CreateSupplier suppliers.CreateSupplier
	ListSuppliers  suppliers.ListSuppliers
	GetSupplier    suppliers.GetSupplier
	UpdateSupplier suppliers.UpdateSupplier
	DeleteSupplier suppliers.DeleteSupplier

	CreateSeller sellers.CreateSeller
	ListSellers  sellers.ListSellers
	GetSeller    sellers.GetSeller
	UpdateSeller sellers.UpdateSeller
	DeleteSeller sellers.DeleteSeller

	ListStockLevels         stock.ListStockLevels
	ListStockMovements      stock.ListStockMovements
	RegisterStockEntry      stock.RegisterStockEntry
	RegisterStockAdjustment stock.RegisterStockAdjustment
	RegisterInitialStock    stock.RegisterInitialStock

	CreateSale sales.CreateSale
	ListSales  sales.ListSales
	GetSale    sales.GetSale
	VoidSale   sales.VoidSale
}

type repositories struct {
	categories gateway.CategoryRepository
	items      gateway.ItemRepository
	customers  gateway.CustomerRepository
	suppliers  gateway.SupplierRepository
	sellers    gateway.SellerRepository
	stock      gateway.StockRepository
	sales      gateway.SaleRepository
	tx         gateway.TxManager
}

func New(cfg environments.Server) (*Container, error) {
	repos, err := newRepositories(cfg.StorageDriver)
	if err != nil {
		return nil, err
	}

	return &Container{
		CreateCategory: categories.NewCreateCategory(repos.categories),
		ListCategories: categories.NewListCategories(repos.categories),
		GetCategory:    categories.NewGetCategory(repos.categories),
		UpdateCategory: categories.NewUpdateCategory(repos.categories),
		DeleteCategory: categories.NewDeleteCategory(repos.categories),

		CreateItem: items.NewCreateItem(repos.items),
		ListItems:  items.NewListItems(repos.items),
		GetItem:    items.NewGetItem(repos.items),
		UpdateItem: items.NewUpdateItem(repos.items),
		DeleteItem: items.NewDeleteItem(repos.items),

		CreateCustomer: customers.NewCreateCustomer(repos.customers),
		ListCustomers:  customers.NewListCustomers(repos.customers),
		GetCustomer:    customers.NewGetCustomer(repos.customers),
		UpdateCustomer: customers.NewUpdateCustomer(repos.customers),
		DeleteCustomer: customers.NewDeleteCustomer(repos.customers),

		CreateSupplier: suppliers.NewCreateSupplier(repos.suppliers),
		ListSuppliers:  suppliers.NewListSuppliers(repos.suppliers),
		GetSupplier:    suppliers.NewGetSupplier(repos.suppliers),
		UpdateSupplier: suppliers.NewUpdateSupplier(repos.suppliers),
		DeleteSupplier: suppliers.NewDeleteSupplier(repos.suppliers),

		CreateSeller: sellers.NewCreateSeller(repos.sellers),
		ListSellers:  sellers.NewListSellers(repos.sellers),
		GetSeller:    sellers.NewGetSeller(repos.sellers),
		UpdateSeller: sellers.NewUpdateSeller(repos.sellers),
		DeleteSeller: sellers.NewDeleteSeller(repos.sellers),

		ListStockLevels:         stock.NewListStockLevels(repos.stock),
		ListStockMovements:      stock.NewListStockMovements(repos.stock),
		RegisterStockEntry:      stock.NewRegisterStockEntry(repos.tx, repos.stock, repos.items, repos.suppliers),
		RegisterStockAdjustment: stock.NewRegisterStockAdjustment(repos.tx, repos.stock, repos.items),
		RegisterInitialStock:    stock.NewRegisterInitialStock(repos.tx, repos.stock, repos.items),

		CreateSale: sales.NewCreateSale(repos.tx, repos.sales, repos.items, repos.customers, repos.sellers, repos.stock, domain.StockPolicy{AllowNegative: cfg.StockAllowNegative}),
		ListSales:  sales.NewListSales(repos.sales),
		GetSale:    sales.NewGetSale(repos.sales),
		VoidSale:   sales.NewVoidSale(repos.tx, repos.sales, repos.stock),
	}, nil
}

func newRepositories(driver environments.StorageDriver) (repositories, error) {
	switch driver {
	case environments.StorageMemory:
		store := memory.NewStore()
		return repositories{
			categories: memory.NewCategoryRepository(store),
			items:      memory.NewItemRepository(store),
			customers:  memory.NewCustomerRepository(store),
			suppliers:  memory.NewSupplierRepository(store),
			sellers:    memory.NewSellerRepository(store),
			stock:      memory.NewStockRepository(store),
			sales:      memory.NewSaleRepository(store),
			tx:         store,
		}, nil
	default:
		return repositories{}, fmt.Errorf("unsupported storage driver %q", driver)
	}
}

// Handlers wraps every use case in its route's handler.
func (container *Container) Handlers() route.Handlers {
	return route.Handlers{
		ListCategories: handler.NewListCategoriesHandler(container.ListCategories),
		GetCategory:    handler.NewGetCategoryHandler(container.GetCategory),
		CreateCategory: handler.NewCreateCategoryHandler(container.CreateCategory),
		UpdateCategory: handler.NewUpdateCategoryHandler(container.UpdateCategory),
		DeleteCategory: handler.NewDeleteCategoryHandler(container.DeleteCategory),

		ListItems:  handler.NewListItemsHandler(container.ListItems),
		GetItem:    handler.NewGetItemHandler(container.GetItem),
		CreateItem: handler.NewCreateItemHandler(container.CreateItem),
		UpdateItem: handler.NewUpdateItemHandler(container.UpdateItem),
		DeleteItem: handler.NewDeleteItemHandler(container.DeleteItem),

		ListCustomers:  handler.NewListCustomersHandler(container.ListCustomers),
		GetCustomer:    handler.NewGetCustomerHandler(container.GetCustomer),
		CreateCustomer: handler.NewCreateCustomerHandler(container.CreateCustomer),
		UpdateCustomer: handler.NewUpdateCustomerHandler(container.UpdateCustomer),
		DeleteCustomer: handler.NewDeleteCustomerHandler(container.DeleteCustomer),

		ListSuppliers:  handler.NewListSuppliersHandler(container.ListSuppliers),
		GetSupplier:    handler.NewGetSupplierHandler(container.GetSupplier),
		CreateSupplier: handler.NewCreateSupplierHandler(container.CreateSupplier),
		UpdateSupplier: handler.NewUpdateSupplierHandler(container.UpdateSupplier),
		DeleteSupplier: handler.NewDeleteSupplierHandler(container.DeleteSupplier),

		ListSellers:  handler.NewListSellersHandler(container.ListSellers),
		GetSeller:    handler.NewGetSellerHandler(container.GetSeller),
		CreateSeller: handler.NewCreateSellerHandler(container.CreateSeller),
		UpdateSeller: handler.NewUpdateSellerHandler(container.UpdateSeller),
		DeleteSeller: handler.NewDeleteSellerHandler(container.DeleteSeller),

		ListStockLevels:       handler.NewListStockLevelsHandler(container.ListStockLevels),
		ListStockMovements:    handler.NewListStockMovementsHandler(container.ListStockMovements),
		CreateStockEntry:      handler.NewCreateStockEntryHandler(container.RegisterStockEntry),
		CreateStockAdjustment: handler.NewCreateStockAdjustmentHandler(container.RegisterStockAdjustment),
		CreateStockInitial:    handler.NewCreateStockInitialHandler(container.RegisterInitialStock),

		ListSales:  handler.NewListSalesHandler(container.ListSales),
		GetSale:    handler.NewGetSaleHandler(container.GetSale),
		CreateSale: handler.NewCreateSaleHandler(container.CreateSale),
		VoidSale:   handler.NewVoidSaleHandler(container.VoidSale),
	}
}
