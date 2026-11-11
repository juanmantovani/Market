package route_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"market/backend/internal/business/domain"
	"market/backend/internal/infrastructure/delivery/webapp/dependencies"
	"market/backend/internal/infrastructure/delivery/webapp/middleware"
	"market/backend/internal/infrastructure/delivery/webapp/route"
	"market/backend/internal/infrastructure/environments"
)

func newMux(t *testing.T) http.Handler {
	t.Helper()
	container, err := dependencies.New(environments.Server{StorageDriver: environments.StorageMemory})
	if err != nil {
		t.Fatalf("dependencies.New() error = %v", err)
	}

	mux := http.NewServeMux()
	route.RegisterRoutes(mux, container.Handlers())

	admin := domain.Actor{ID: "admin", Role: domain.RoleAdmin}
	return middleware.DevActor(admin, mux)
}

func TestRegisterRoutes(t *testing.T) {
	mux := newMux(t)

	created := httptest.NewRecorder()
	mux.ServeHTTP(created, httptest.NewRequest(http.MethodPost, "/api/v1/categories", strings.NewReader(`{"name":"Bebidas"}`)))
	if created.Code != http.StatusCreated {
		t.Fatalf("POST /api/v1/categories = %d, body %s", created.Code, created.Body.String())
	}

	tests := []struct {
		method string
		path   string
		body   string
		want   int
	}{
		{method: http.MethodGet, path: "/health/live", want: http.StatusOK},
		{method: http.MethodPost, path: "/health/live", want: http.StatusMethodNotAllowed},
		{method: http.MethodGet, path: "/api/v1/unknown", want: http.StatusNotFound},
		{method: http.MethodGet, path: "/api/v1/categories", want: http.StatusOK},
		{method: http.MethodGet, path: "/api/v1/categories/missing", want: http.StatusNotFound},
		{method: http.MethodPatch, path: "/api/v1/categories/missing", body: `{"name":"X"}`, want: http.StatusNotFound},
		{method: http.MethodDelete, path: "/api/v1/categories/missing", want: http.StatusNotFound},
		{method: http.MethodGet, path: "/api/v1/items", want: http.StatusOK},
		{method: http.MethodPost, path: "/api/v1/items", body: `{"code":"X"}`, want: http.StatusBadRequest},
		{method: http.MethodGet, path: "/api/v1/items/missing", want: http.StatusNotFound},
		{method: http.MethodPut, path: "/api/v1/items/missing", body: `{}`, want: http.StatusNotFound},
		{method: http.MethodDelete, path: "/api/v1/items/missing", want: http.StatusNotFound},
		{method: http.MethodGet, path: "/api/v1/customers", want: http.StatusOK},
		{method: http.MethodPost, path: "/api/v1/customers", body: `{"name":""}`, want: http.StatusBadRequest},
		{method: http.MethodGet, path: "/api/v1/customers/missing", want: http.StatusNotFound},
		{method: http.MethodPut, path: "/api/v1/customers/missing", body: `{}`, want: http.StatusNotFound},
		{method: http.MethodDelete, path: "/api/v1/customers/missing", want: http.StatusNotFound},
		{method: http.MethodGet, path: "/api/v1/suppliers", want: http.StatusOK},
		{method: http.MethodPost, path: "/api/v1/suppliers", body: `{"businessName":""}`, want: http.StatusBadRequest},
		{method: http.MethodGet, path: "/api/v1/suppliers/missing", want: http.StatusNotFound},
		{method: http.MethodPut, path: "/api/v1/suppliers/missing", body: `{}`, want: http.StatusNotFound},
		{method: http.MethodDelete, path: "/api/v1/suppliers/missing", want: http.StatusNotFound},
		// The /api/ catch-all answers unsupported methods with the JSON 404
		// instead of the mux's plain-text 405, so every API error shares one
		// shape.
		{method: http.MethodPut, path: "/api/v1/categories/missing", want: http.StatusNotFound},
	}

	for _, test := range tests {
		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, httptest.NewRequest(test.method, test.path, strings.NewReader(test.body)))

		if recorder.Code != test.want {
			t.Errorf("%s %s = %d, want %d", test.method, test.path, recorder.Code, test.want)
		}
	}
}
