package route

import (
	"net/http"
	"time"

	"market/backend/internal/infrastructure/delivery/webapp/handler"
)

const (
	catalogTimeout = 5 * time.Second
	stockTimeout   = 5 * time.Second
	salesTimeout   = 5 * time.Second
)

// MaxHandlerTimeout is the longest timeout of any route registered here.
// The HTTP server's deadlines derive from it, so a handler that runs to its
// limit can still write its response.
const MaxHandlerTimeout = max(catalogTimeout, stockTimeout, salesTimeout)

// Handlers holds one handler per route. Each feature adds its own.
type Handlers struct {
	ListCategories *handler.ListCategoriesHandler
	GetCategory    *handler.GetCategoryHandler
	CreateCategory *handler.CreateCategoryHandler
	UpdateCategory *handler.UpdateCategoryHandler
	DeleteCategory *handler.DeleteCategoryHandler

	ListItems  *handler.ListItemsHandler
	GetItem    *handler.GetItemHandler
	CreateItem *handler.CreateItemHandler
	UpdateItem *handler.UpdateItemHandler
	DeleteItem *handler.DeleteItemHandler

	ListCustomers  *handler.ListCustomersHandler
	GetCustomer    *handler.GetCustomerHandler
	CreateCustomer *handler.CreateCustomerHandler
	UpdateCustomer *handler.UpdateCustomerHandler
	DeleteCustomer *handler.DeleteCustomerHandler

	ListSuppliers  *handler.ListSuppliersHandler
	GetSupplier    *handler.GetSupplierHandler
	CreateSupplier *handler.CreateSupplierHandler
	UpdateSupplier *handler.UpdateSupplierHandler
	DeleteSupplier *handler.DeleteSupplierHandler

	ListSellers  *handler.ListSellersHandler
	GetSeller    *handler.GetSellerHandler
	CreateSeller *handler.CreateSellerHandler
	UpdateSeller *handler.UpdateSellerHandler
	DeleteSeller *handler.DeleteSellerHandler

	ListStockLevels       *handler.ListStockLevelsHandler
	ListStockMovements    *handler.ListStockMovementsHandler
	CreateStockEntry      *handler.CreateStockEntryHandler
	CreateStockAdjustment *handler.CreateStockAdjustmentHandler

	CreateSale *handler.CreateSaleHandler
}

func RegisterRoutes(mux *http.ServeMux, handlers Handlers) {
	mux.HandleFunc("GET /health/live", handler.Live)
	mux.HandleFunc("/api/", handler.NotFound)

	mux.Handle("GET /api/v1/categories", handler.Adapt(handlers.ListCategories, catalogTimeout))
	mux.Handle("POST /api/v1/categories", handler.Adapt(handlers.CreateCategory, catalogTimeout))
	mux.Handle("GET /api/v1/categories/{id}", handler.Adapt(handlers.GetCategory, catalogTimeout))
	mux.Handle("PATCH /api/v1/categories/{id}", handler.Adapt(handlers.UpdateCategory, catalogTimeout))
	mux.Handle("DELETE /api/v1/categories/{id}", handler.Adapt(handlers.DeleteCategory, catalogTimeout))

	mux.Handle("GET /api/v1/items", handler.Adapt(handlers.ListItems, catalogTimeout))
	mux.Handle("POST /api/v1/items", handler.Adapt(handlers.CreateItem, catalogTimeout))
	mux.Handle("GET /api/v1/items/{id}", handler.Adapt(handlers.GetItem, catalogTimeout))
	mux.Handle("PUT /api/v1/items/{id}", handler.Adapt(handlers.UpdateItem, catalogTimeout))
	mux.Handle("DELETE /api/v1/items/{id}", handler.Adapt(handlers.DeleteItem, catalogTimeout))

	mux.Handle("GET /api/v1/customers", handler.Adapt(handlers.ListCustomers, catalogTimeout))
	mux.Handle("POST /api/v1/customers", handler.Adapt(handlers.CreateCustomer, catalogTimeout))
	mux.Handle("GET /api/v1/customers/{id}", handler.Adapt(handlers.GetCustomer, catalogTimeout))
	mux.Handle("PUT /api/v1/customers/{id}", handler.Adapt(handlers.UpdateCustomer, catalogTimeout))
	mux.Handle("DELETE /api/v1/customers/{id}", handler.Adapt(handlers.DeleteCustomer, catalogTimeout))

	mux.Handle("GET /api/v1/suppliers", handler.Adapt(handlers.ListSuppliers, catalogTimeout))
	mux.Handle("POST /api/v1/suppliers", handler.Adapt(handlers.CreateSupplier, catalogTimeout))
	mux.Handle("GET /api/v1/suppliers/{id}", handler.Adapt(handlers.GetSupplier, catalogTimeout))
	mux.Handle("PUT /api/v1/suppliers/{id}", handler.Adapt(handlers.UpdateSupplier, catalogTimeout))
	mux.Handle("DELETE /api/v1/suppliers/{id}", handler.Adapt(handlers.DeleteSupplier, catalogTimeout))

	mux.Handle("GET /api/v1/sellers", handler.Adapt(handlers.ListSellers, catalogTimeout))
	mux.Handle("POST /api/v1/sellers", handler.Adapt(handlers.CreateSeller, catalogTimeout))
	mux.Handle("GET /api/v1/sellers/{id}", handler.Adapt(handlers.GetSeller, catalogTimeout))
	mux.Handle("PUT /api/v1/sellers/{id}", handler.Adapt(handlers.UpdateSeller, catalogTimeout))
	mux.Handle("DELETE /api/v1/sellers/{id}", handler.Adapt(handlers.DeleteSeller, catalogTimeout))

	mux.Handle("GET /api/v1/stock", handler.Adapt(handlers.ListStockLevels, stockTimeout))
	mux.Handle("GET /api/v1/stock/movements", handler.Adapt(handlers.ListStockMovements, stockTimeout))
	mux.Handle("POST /api/v1/stock/entries", handler.Adapt(handlers.CreateStockEntry, stockTimeout))
	mux.Handle("POST /api/v1/stock/adjustments", handler.Adapt(handlers.CreateStockAdjustment, stockTimeout))

	mux.Handle("POST /api/v1/sales", handler.Adapt(handlers.CreateSale, salesTimeout))
}
