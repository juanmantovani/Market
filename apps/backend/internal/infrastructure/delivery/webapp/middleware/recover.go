package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"market/backend/internal/infrastructure/delivery/webapp/response"
)

// Recover turns a panicking handler into a generic 500 and logs the panic
// with its stack, instead of letting net/http drop the connection.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}
			if recovered == http.ErrAbortHandler {
				panic(recovered)
			}
			err := fmt.Errorf("panic: %v\n%s", recovered, debug.Stack())
			response.WriteError(w, request, "handler panicked: "+request.Method+" "+request.URL.Path, err)
		}()

		next.ServeHTTP(w, request)
	})
}
