package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewDerivesDeadlinesFromTheLongestHandler(t *testing.T) {
	handler := http.NewServeMux()

	server := New("8080", handler, 5*time.Second)

	if server.Addr != ":8080" {
		t.Errorf("Addr = %q, want :8080", server.Addr)
	}
	if server.Handler != handler {
		t.Error("Handler was not wired")
	}
	if server.ReadTimeout != 15*time.Second {
		t.Errorf("ReadTimeout = %v, want 15s", server.ReadTimeout)
	}
	if server.WriteTimeout != 25*time.Second {
		t.Errorf("WriteTimeout = %v, want 25s", server.WriteTimeout)
	}
	if server.ReadHeaderTimeout == 0 || server.IdleTimeout == 0 {
		t.Error("header and idle timeouts must be set")
	}
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestWithCORSWithoutOriginReturnsTheHandlerUntouched(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	request.Header.Set("Origin", "http://evil.test")

	WithCORS("", okHandler()).ServeHTTP(recorder, request)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}

func TestWithCORSAllowsOnlyTheConfiguredOrigin(t *testing.T) {
	const allowed = "http://localhost:5173"
	handler := WithCORS(allowed, okHandler())

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	request.Header.Set("Origin", allowed)
	handler.ServeHTTP(recorder, request)
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != allowed {
		t.Fatalf("allowed origin: Access-Control-Allow-Origin = %q, want %q", got, allowed)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("allowed origin: status = %d, want 200", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	request.Header.Set("Origin", "http://evil.test")
	handler.ServeHTTP(recorder, request)
	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("other origin: Access-Control-Allow-Origin = %q, want empty", got)
	}
}

func TestWithCORSAnswersPreflight(t *testing.T) {
	const allowed = "http://localhost:5173"
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/sales", nil)
	request.Header.Set("Origin", allowed)

	WithCORS(allowed, okHandler()).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", recorder.Code)
	}
	if got := recorder.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Fatal("preflight did not list the allowed headers")
	}
}
