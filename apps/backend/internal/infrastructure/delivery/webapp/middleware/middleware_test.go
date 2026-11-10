package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"market/backend/internal/business/domain"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

func TestDevActorExposesTheActorToHandlers(t *testing.T) {
	want := domain.Actor{ID: "dev", Name: "Desarrollo", Role: domain.RoleAdmin}
	var got domain.Actor
	var found bool
	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got, found = ActorFromContext(r.Context())
	})

	DevActor(want, next).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	if !found || got != want {
		t.Fatalf("ActorFromContext() = %+v, %v; want %+v, true", got, found, want)
	}
}

func TestActorFromContextWithoutActor(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	if _, found := ActorFromContext(request.Context()); found {
		t.Fatal("ActorFromContext() found an actor in a bare context")
	}
}

func TestRecoverAnswersAPanicWithAGeneric500(t *testing.T) {
	recorder := httptest.NewRecorder()
	panicking := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})

	Recover(panicking).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/items", nil))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", recorder.Code)
	}
	var body response.ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("decoding body: %v", err)
	}
	if body.Code != "internal_error" || strings.Contains(body.Message, "boom") {
		t.Fatalf("body = %+v, want a generic internal_error", body)
	}
}

func TestRecoverPassesThroughWithoutPanic(t *testing.T) {
	recorder := httptest.NewRecorder()
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	Recover(ok).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", recorder.Code)
	}
}

func TestRecoverRepanicsOnAbortHandler(t *testing.T) {
	aborting := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	})

	defer func() {
		if recovered := recover(); recovered != http.ErrAbortHandler {
			t.Fatalf("recovered %v, want http.ErrAbortHandler", recovered)
		}
	}()

	Recover(aborting).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
}

func TestRequestLogRecordsMethodPathAndStatus(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	teapot := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	recorder := httptest.NewRecorder()

	RequestLog(teapot).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/sales", nil))

	if recorder.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want 418", recorder.Code)
	}
	line := logs.String()
	for _, want := range []string{"method=POST", "path=/api/v1/sales", "status=418"} {
		if !strings.Contains(line, want) {
			t.Errorf("log %q does not contain %q", line, want)
		}
	}
}

func TestStatusRecorderUnwraps(t *testing.T) {
	underlying := httptest.NewRecorder()
	recorder := &statusRecorder{ResponseWriter: underlying}

	if recorder.Unwrap() != underlying {
		t.Fatal("Unwrap() did not return the underlying writer")
	}
}
