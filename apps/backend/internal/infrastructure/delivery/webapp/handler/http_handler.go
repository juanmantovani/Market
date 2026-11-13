package handler

import (
	"context"
	"net/http"
	"time"

	"market/backend/internal/infrastructure/delivery/webapp/response"
)

// MaxBodyBytes caps every request body. The API only receives JSON forms,
// so anything larger is a mistake or abuse.
const MaxBodyBytes = 1 << 20

// HTTPHandler is a single route's handler. It returns an error instead of
// writing one, so every route's failure path goes through
// response.WriteError.
type HTTPHandler interface {
	Handle(w http.ResponseWriter, r *http.Request) error
}

// Adapt turns an HTTPHandler into an http.HandlerFunc, bounding the request
// context to timeout and the body to MaxBodyBytes so no handler sets those up
// on its own.
func Adapt(h HTTPHandler, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		if err := h.Handle(w, r.WithContext(ctx)); err != nil {
			response.WriteError(w, r, "request failed: "+r.Method+" "+r.URL.Path, err)
		}
	}
}
