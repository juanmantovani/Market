package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	dto "market/backend/internal/infrastructure/delivery/webapp/dto/health"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

func TestLive(t *testing.T) {
	recorder := httptest.NewRecorder()

	Live(recorder, httptest.NewRequest(http.MethodGet, "/health/live", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if got := decodeBody[dto.StatusResponse](t, recorder); got.Status != "ok" {
		t.Fatalf("status field = %q, want ok", got.Status)
	}
}

func TestReadyRespondsOkWhenStorageIsReachable(t *testing.T) {
	recorder := httptest.NewRecorder()
	handler := NewReadyHandler(func(context.Context) error { return nil })

	if err := handler.Handle(recorder, httptest.NewRequest(http.MethodGet, "/health/ready", nil)); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	if got := decodeBody[dto.StatusResponse](t, recorder); got.Status != "ok" {
		t.Fatalf("status field = %q, want ok", got.Status)
	}
}

func TestReadyFailsWhenStorageIsUnreachable(t *testing.T) {
	recorder := httptest.NewRecorder()
	cause := errors.New("connection refused")
	handler := NewReadyHandler(func(context.Context) error { return cause })

	err := handler.Handle(recorder, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if !errors.Is(err, cause) {
		t.Fatalf("Handle() error = %v, want it to wrap %v", err, cause)
	}
}

func TestNotFound(t *testing.T) {
	recorder := httptest.NewRecorder()

	NotFound(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/nope", nil))

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
	if got := decodeBody[response.ErrorResponse](t, recorder); got.Code != "route_not_found" {
		t.Fatalf("code = %q, want route_not_found", got.Code)
	}
}
