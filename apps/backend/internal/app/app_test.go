package app

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

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

	application, err := New()
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

	if _, err := New(); err == nil {
		t.Fatal("New() error = nil, want a configuration error")
	}
}

func TestRunShutsDownWhenTheContextIsCancelled(t *testing.T) {
	t.Setenv("PORT", freePort(t))
	t.Setenv("STORAGE_DRIVER", "memory")

	application, err := New()
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
