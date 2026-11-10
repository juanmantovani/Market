package handler

import (
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
