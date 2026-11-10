package route

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterRoutes(t *testing.T) {
	mux := http.NewServeMux()
	RegisterRoutes(mux)

	tests := []struct {
		method string
		path   string
		want   int
	}{
		{method: http.MethodGet, path: "/health/live", want: http.StatusOK},
		{method: http.MethodPost, path: "/health/live", want: http.StatusMethodNotAllowed},
		{method: http.MethodGet, path: "/api/v1/unknown", want: http.StatusNotFound},
	}

	for _, test := range tests {
		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, httptest.NewRequest(test.method, test.path, nil))

		if recorder.Code != test.want {
			t.Errorf("%s %s = %d, want %d", test.method, test.path, recorder.Code, test.want)
		}
	}
}
