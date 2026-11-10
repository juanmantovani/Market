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

	deleted := httptest.NewRecorder()
	api.ServeHTTP(deleted, httptest.NewRequest(http.MethodDelete, "/api/v1/categories/"+page.Items[0].ID, nil))
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("DELETE = %d %s", deleted.Code, deleted.Body.String())
	}

	gone := httptest.NewRecorder()
	api.ServeHTTP(gone, httptest.NewRequest(http.MethodGet, "/api/v1/categories/"+page.Items[0].ID, nil))
	if gone.Code != http.StatusNotFound {
		t.Fatalf("GET after DELETE = %d, want 404", gone.Code)
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
