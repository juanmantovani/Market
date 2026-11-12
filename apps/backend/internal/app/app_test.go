package app

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestNewSeedsDemoDataAndServesTheCategoryAPI(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("STORAGE_DRIVER", "memory")
	t.Setenv("SEED_DEMO_DATA", "true")

	application, err := New(context.Background())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	api := application.server.Handler

	listed := httptest.NewRecorder()
	api.ServeHTTP(listed, httptest.NewRequest(http.MethodGet, "/api/v1/categories?q=bebidas", nil))
	var page struct {
		Items []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"items"`
		Total int `json:"total"`
	}
	if err := json.NewDecoder(listed.Body).Decode(&page); err != nil || listed.Code != http.StatusOK {
		t.Fatalf("GET categories = %d, %v", listed.Code, err)
	}
	if page.Total != 1 || page.Items[0].Name != "Bebidas" {
		t.Fatalf("search for bebidas = %+v", page)
	}

	duplicate := httptest.NewRecorder()
	api.ServeHTTP(duplicate, httptest.NewRequest(http.MethodPost, "/api/v1/categories", strings.NewReader(`{"name":"BEBIDAS"}`)))
	if duplicate.Code != http.StatusConflict {
		t.Fatalf("POST duplicate = %d, want 409", duplicate.Code)
	}

	bebidasItems := httptest.NewRecorder()
	api.ServeHTTP(bebidasItems, httptest.NewRequest(http.MethodGet, "/api/v1/items?categoryId="+page.Items[0].ID, nil))
	if bebidasItems.Code != http.StatusOK || !strings.Contains(bebidasItems.Body.String(), `"total":2`) {
		t.Fatalf("GET items of Bebidas = %d %s", bebidasItems.Code, bebidasItems.Body.String())
	}

	inUse := httptest.NewRecorder()
	api.ServeHTTP(inUse, httptest.NewRequest(http.MethodDelete, "/api/v1/categories/"+page.Items[0].ID, nil))
	if inUse.Code != http.StatusConflict || !strings.Contains(inUse.Body.String(), "category_in_use") {
		t.Fatalf("DELETE category with items = %d %s, want 409 category_in_use", inUse.Code, inUse.Body.String())
	}

	created := httptest.NewRecorder()
	api.ServeHTTP(created, httptest.NewRequest(http.MethodPost, "/api/v1/categories", strings.NewReader(`{"name":"Ferretería"}`)))
	var ferreteria struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(created.Body).Decode(&ferreteria); err != nil || created.Code != http.StatusCreated {
		t.Fatalf("POST category = %d, %v", created.Code, err)
	}

	deleted := httptest.NewRecorder()
	api.ServeHTTP(deleted, httptest.NewRequest(http.MethodDelete, "/api/v1/categories/"+ferreteria.ID, nil))
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("DELETE empty category = %d %s", deleted.Code, deleted.Body.String())
	}

	gone := httptest.NewRecorder()
	api.ServeHTTP(gone, httptest.NewRequest(http.MethodGet, "/api/v1/categories/"+ferreteria.ID, nil))
	if gone.Code != http.StatusNotFound {
		t.Fatalf("GET after DELETE = %d, want 404", gone.Code)
	}
}

func TestStockFlowOverTheDemoData(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("STORAGE_DRIVER", "memory")
	t.Setenv("SEED_DEMO_DATA", "true")

	application, err := New(context.Background())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	call := func(method, target, body string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		application.server.Handler.ServeHTTP(recorder, httptest.NewRequest(method, target, strings.NewReader(body)))
		return recorder
	}
	type level struct {
		ItemID   string `json:"itemId"`
		Code     string `json:"code"`
		Quantity int64  `json:"quantity"`
	}
	lowStock := func() []level {
		recorder := call(http.MethodGet, "/api/v1/stock?lowStock=true", "")
		var page struct {
			Items []level `json:"items"`
		}
		if err := json.NewDecoder(recorder.Body).Decode(&page); err != nil || recorder.Code != http.StatusOK {
			t.Fatalf("GET low stock = %d, %v", recorder.Code, err)
		}
		return page.Items
	}

	low := lowStock()
	if len(low) != 3 || low[2].Code != "ALM-001" || low[2].Quantity != 4_000 {
		t.Fatalf("low stock = %+v, want leche, queso and yerba", low)
	}
	yerba := low[2].ItemID

	entry := call(http.MethodPost, "/api/v1/stock/entries", `{"lines":[{"itemId":"`+yerba+`","quantity":12000}]}`)
	if entry.Code != http.StatusCreated || !strings.Contains(entry.Body.String(), `"balanceAfter":16000`) {
		t.Fatalf("POST entry = %d %s", entry.Code, entry.Body.String())
	}
	if after := lowStock(); len(after) != 2 {
		t.Fatalf("low stock after the entry = %+v, want yerba gone", after)
	}

	tooMuch := call(http.MethodPost, "/api/v1/stock/adjustments", `{"itemId":"`+yerba+`","quantity":-17000,"reason":"Conteo"}`)
	if tooMuch.Code != http.StatusConflict || !strings.Contains(tooMuch.Body.String(), "insufficient_stock") {
		t.Fatalf("POST adjustment below zero = %d %s", tooMuch.Code, tooMuch.Body.String())
	}
	adjusted := call(http.MethodPost, "/api/v1/stock/adjustments", `{"itemId":"`+yerba+`","quantity":-1000,"reason":"Rotura"}`)
	if adjusted.Code != http.StatusCreated || !strings.Contains(adjusted.Body.String(), `"balanceAfter":15000`) {
		t.Fatalf("POST adjustment = %d %s", adjusted.Code, adjusted.Body.String())
	}

	history := call(http.MethodGet, "/api/v1/stock/movements?itemId="+yerba, "")
	var movements struct {
		Items []struct {
			Type string `json:"type"`
		} `json:"items"`
		Total int `json:"total"`
	}
	if err := json.NewDecoder(history.Body).Decode(&movements); err != nil || movements.Total != 3 {
		t.Fatalf("GET movements = %+v, %v", movements, err)
	}
	if movements.Items[0].Type != "adjustment" || movements.Items[1].Type != "entry" || movements.Items[2].Type != "initial" {
		t.Fatalf("movements = %+v, want newest first", movements.Items)
	}

	inUse := call(http.MethodDelete, "/api/v1/items/"+yerba, "")
	if inUse.Code != http.StatusConflict || !strings.Contains(inUse.Body.String(), "item_in_use") {
		t.Fatalf("DELETE item with movements = %d %s", inUse.Code, inUse.Body.String())
	}
}

func TestNewStartsEmptyWithoutDemoData(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("STORAGE_DRIVER", "memory")
	t.Setenv("SEED_DEMO_DATA", "false")

	application, err := New(context.Background())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	application.server.Handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil))
	if !strings.Contains(recorder.Body.String(), `"total":0`) {
		t.Fatalf("GET categories without demo data = %s", recorder.Body.String())
	}
}

func freePort(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("finding a free port: %v", err)
	}
	defer listener.Close()

	return strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
}

func TestNewServesLivenessThroughTheMiddlewareChain(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("STORAGE_DRIVER", "memory")
	t.Setenv("CORS_ALLOWED_ORIGIN", "http://localhost:5173")

	application, err := New(context.Background())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	application.server.Handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if recorder.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatal("CORS middleware is not wired")
	}
}

func TestNewRejectsInvalidConfiguration(t *testing.T) {
	t.Setenv("STORAGE_DRIVER", "sqlite")

	if _, err := New(context.Background()); err == nil {
		t.Fatal("New() error = nil, want a configuration error")
	}
}

func TestRunShutsDownWhenTheContextIsCancelled(t *testing.T) {
	t.Setenv("PORT", freePort(t))
	t.Setenv("STORAGE_DRIVER", "memory")

	application, err := New(context.Background())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- application.Run(ctx) }()

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() error = %v, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run() did not return after cancellation")
	}
}
