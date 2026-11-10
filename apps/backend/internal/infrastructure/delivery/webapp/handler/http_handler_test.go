package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type handlerFunc func(w http.ResponseWriter, r *http.Request) error

func (fn handlerFunc) Handle(w http.ResponseWriter, r *http.Request) error {
	return fn(w, r)
}

type echoRequest struct {
	Name string `json:"name"`
}

func echoHandler() HTTPHandler {
	return handlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		body, err := decodeJSON[echoRequest](r)
		if err != nil {
			return err
		}
		response.WriteJSON(w, http.StatusCreated, body)
		return nil
	})
}

func decodeBody[T any](t *testing.T, recorder *httptest.ResponseRecorder) T {
	t.Helper()

	var body T
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("decoding the response body: %v", err)
	}
	return body
}

func TestAdaptWritesTheHandlerResponse(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/echo", strings.NewReader(`{"name":"Yerba"}`))

	Adapt(echoHandler(), time.Second)(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", recorder.Code)
	}
	if got := decodeBody[echoRequest](t, recorder); got.Name != "Yerba" {
		t.Fatalf("body = %+v", got)
	}
}

func TestAdaptTranslatesTheReturnedError(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	failing := handlerFunc(func(http.ResponseWriter, *http.Request) error {
		return domain.ErrForbidden
	})

	Adapt(failing, time.Second)(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", recorder.Code)
	}
	if got := decodeBody[response.ErrorResponse](t, recorder); got.Code != "forbidden" {
		t.Fatalf("code = %q, want forbidden", got.Code)
	}
}

func TestAdaptBoundsTheContextToTheTimeout(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
	var deadline time.Time
	var hasDeadline bool
	probe := handlerFunc(func(_ http.ResponseWriter, r *http.Request) error {
		deadline, hasDeadline = r.Context().Deadline()
		return nil
	})

	const timeout = 2 * time.Second
	before := time.Now()
	Adapt(probe, timeout)(recorder, request)
	after := time.Now()

	if !hasDeadline {
		t.Fatal("handler context has no deadline")
	}
	if deadline.Before(before.Add(timeout)) || deadline.After(after.Add(timeout)) {
		t.Fatalf("deadline %v, want between %v and %v", deadline, before.Add(timeout), after.Add(timeout))
	}
}

func TestDecodeJSONRejectsMalformedBodies(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/echo", strings.NewReader(`{"name":`))

	Adapt(echoHandler(), time.Second)(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}
	if got := decodeBody[response.ErrorResponse](t, recorder); got.Code != "invalid_body" {
		t.Fatalf("code = %q, want invalid_body", got.Code)
	}
}

func TestAdaptRejectsOversizedBodies(t *testing.T) {
	recorder := httptest.NewRecorder()
	oversized := `{"name":"` + strings.Repeat("a", MaxBodyBytes) + `"}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/echo", strings.NewReader(oversized))

	Adapt(echoHandler(), time.Second)(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}
	body := decodeBody[response.ErrorResponse](t, recorder)
	if body.Code != "invalid_body" {
		t.Fatalf("code = %q, want invalid_body", body.Code)
	}
}

func TestDecodeJSONKeepsTheDecoderErrorAsCause(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not json"))

	_, err := decodeJSON[echoRequest](request)

	var domainErr *domain.Error
	if !errors.As(err, &domainErr) || domainErr.Cause == nil {
		t.Fatalf("decodeJSON() error = %v, want a domain error with a cause", err)
	}
}
