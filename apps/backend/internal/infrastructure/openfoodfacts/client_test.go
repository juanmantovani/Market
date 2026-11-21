package openfoodfacts

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"market/backend/internal/business/domain"
)

func server(t *testing.T, handle http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handle)
	t.Cleanup(server.Close)

	return NewClient(server.Client(), server.URL)
}

func TestNewClientDefaultsToTheProductionAPI(t *testing.T) {
	client := NewClient(nil, "")
	if client.httpClient != http.DefaultClient || client.baseURL != BaseURL {
		t.Fatalf("client = %+v", client)
	}
}

func TestLookupReturnsTheSuggestedProduct(t *testing.T) {
	var gotPath, gotUserAgent string
	client := server(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotUserAgent = r.URL.RequestURI(), r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"3017620422003","product":{"product_name":"Nutella","product_name_es":"Nutella ES","brands":"Ferrero, Nutella","quantity":"400 g"},"status":1,"status_verbose":"product found"}`))
	})

	got, err := client.Lookup(context.Background(), "3017620422003")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if got.Name != "Nutella ES" || got.Description == nil || *got.Description != "Ferrero, Nutella · 400 g" {
		t.Fatalf("got = %+v", got)
	}
	if !strings.Contains(gotPath, "/api/v2/product/3017620422003.json") {
		t.Fatalf("path = %q", gotPath)
	}
	if gotUserAgent == "" {
		t.Fatal("Lookup() sent no User-Agent, which Open Food Facts requires")
	}
}

func TestLookupFallsBackToTheDefaultName(t *testing.T) {
	client := server(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"code":"1","product":{"product_name":"Default name"},"status":1}`))
	})

	got, err := client.Lookup(context.Background(), "1")
	if err != nil || got.Name != "Default name" || got.Description != nil {
		t.Fatalf("got = %+v, err = %v", got, err)
	}
}

func TestLookupReturnsNotFound(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{"http 404", http.StatusNotFound, `{"code":"0","status":0,"status_verbose":"product not found"}`},
		{"status 0 with 200", http.StatusOK, `{"code":"0","status":0,"status_verbose":"no code or invalid code"}`},
		{"found but without a name", http.StatusOK, `{"code":"1","product":{"brands":"Mystery"},"status":1}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := server(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			})

			_, err := client.Lookup(context.Background(), "0")
			if !errors.Is(err, domain.ErrExternalProductNotFound) {
				t.Fatalf("Lookup() error = %v, want ErrExternalProductNotFound", err)
			}
		})
	}
}

func TestLookupFailsOnAnUnexpectedResponse(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{"server error", http.StatusInternalServerError, ""},
		{"malformed json", http.StatusOK, "not json"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := server(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			})

			_, err := client.Lookup(context.Background(), "0")
			if err == nil || errors.Is(err, domain.ErrExternalProductNotFound) {
				t.Fatalf("Lookup() error = %v, want a plain error", err)
			}
		})
	}
}

func TestLookupFailsWhenTheServerIsUnreachable(t *testing.T) {
	client := NewClient(http.DefaultClient, "http://127.0.0.1:0")

	_, err := client.Lookup(context.Background(), "0")
	if err == nil || errors.Is(err, domain.ErrExternalProductNotFound) {
		t.Fatalf("Lookup() error = %v, want a plain connection error", err)
	}
}
