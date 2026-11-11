package route

import (
	"net/http"
	"time"

	"market/backend/internal/infrastructure/delivery/webapp/handler"
)

const catalogTimeout = 5 * time.Second

// MaxHandlerTimeout is the longest timeout of any route registered here.
// The HTTP server's deadlines derive from it, so a handler that runs to its
// limit can still write its response; raise it when a route needs more.
const MaxHandlerTimeout = catalogTimeout

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
}
