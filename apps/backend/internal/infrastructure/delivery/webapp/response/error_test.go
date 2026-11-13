package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"market/backend/internal/business/domain"
)

func decodeError(t *testing.T, recorder *httptest.ResponseRecorder) ErrorResponse {
	t.Helper()

	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
	var body ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("decoding the error body: %v", err)
	}
	return body
}

func TestWriteErrorMapsEachKindToAStatus(t *testing.T) {
	tests := []struct {
		kind domain.Kind
		want int
	}{
		{kind: domain.KindInvalid, want: http.StatusBadRequest},
		{kind: domain.KindForbidden, want: http.StatusForbidden},
		{kind: domain.KindConflict, want: http.StatusConflict},
		{kind: domain.KindNotFound, want: http.StatusNotFound},
		{kind: domain.KindUnavailable, want: http.StatusServiceUnavailable},
		{kind: "unknown", want: http.StatusInternalServerError},
	}

	for _, test := range tests {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
		err := &domain.Error{Kind: test.kind, Code: "some_code", Message: "Algo salió mal"}

		WriteError(recorder, request, "request failed", err)

		if recorder.Code != test.want {
			t.Errorf("kind %q: status = %d, want %d", test.kind, recorder.Code, test.want)
		}
		body := decodeError(t, recorder)
		if body.Code != "some_code" || body.Message != "Algo salió mal" {
			t.Errorf("kind %q: body = %+v", test.kind, body)
		}
	}
}

func TestWriteErrorHidesTheCause(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	err := domain.ErrStorageUnavailable.WithCause(errors.New("secret connection string"))

	WriteError(recorder, request, "request failed", err)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", recorder.Code)
	}
	body := decodeError(t, recorder)
	if body.Code != "storage_unavailable" || body.Message != domain.ErrStorageUnavailable.Message {
		t.Fatalf("body = %+v", body)
	}
}

func TestWriteErrorHidesUnexpectedErrors(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)

	WriteError(recorder, request, "request failed", errors.New("nil pointer somewhere"))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", recorder.Code)
	}
	body := decodeError(t, recorder)
	if body.Code != "internal_error" {
		t.Fatalf("code = %q, want internal_error", body.Code)
	}
}

func TestWriteJSONLogsAnUnencodablePayload(t *testing.T) {
	recorder := httptest.NewRecorder()

	WriteJSON(recorder, http.StatusOK, map[string]any{"bad": make(chan int)})

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
}
