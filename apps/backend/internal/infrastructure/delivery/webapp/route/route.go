package route

import (
	"net/http"
	"time"

	"market/backend/internal/infrastructure/delivery/webapp/handler"
)

// MaxHandlerTimeout is the longest timeout of any route registered here.
// The HTTP server's deadlines derive from it, so a handler that runs to its
// limit can still write its response; raise it when a route needs more.
const MaxHandlerTimeout = 5 * time.Second

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health/live", handler.Live)
	mux.HandleFunc("/api/", handler.NotFound)
}
