package gs1argentina

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

const successPage = `
<form id="frmGtin"></form>
<div class="alert alert-success"></div>
<p id="Descripcion"><b>Descripción: </b>GASEOSA</p>
<p id="marca"><b>Marca: </b>COCA COLA</p>
<p id="variedad"><b>Variedad: </b>Cola Regular</p>
<p id="contenidoneto"><b>Contenido Neto: </b>2,00</p>
<p id="unidadmedida"><b>Unidad Medida: </b>(LT) LITROS</p>
<p id="envaseg"><b>Envase: </b>(C7) PET</p>
<p id="Empresa"><b>Nombre Empresa: </b>ACME</p>
<p id="codigointerno"><b>Código Interno: </b>-</p>
`

func portal(t *testing.T, handle http.Handler) *Client {
	t.Helper()
	server := httptest.NewServer(handle)
	t.Cleanup(server.Close)

	return NewClient(server.Client(), server.URL, "member", "secret", nil)
}

func TestNewClientDefaults(t *testing.T) {
	client := NewClient(nil, "", "", "", nil)
	if client.baseURL != BaseURL || client.httpClient.Timeout == 0 || client.httpClient.Jar == nil {
		t.Fatalf("client = %+v", client)
	}
}

func TestLookupSignsInAndReadsTheProduct(t *testing.T) {
	var logins atomic.Int32
	client := portal(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case loginPath:
			logins.Add(1)
			if r.FormValue("NombreUsuario") != "member" || r.FormValue("Password") != "secret" {
				t.Errorf("login form = %s / %s", r.FormValue("NombreUsuario"), r.FormValue("Password"))
			}
			if r.Header.Get("User-Agent") != userAgent {
				t.Errorf("User-Agent = %q", r.Header.Get("User-Agent"))
			}
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "ok", Path: "/"})
			_, _ = w.Write([]byte(`{"Status":"OK"}`))
		case lookupPath:
			if _, err := r.Cookie("session"); err != nil {
				t.Errorf("lookup without session: %v", err)
			}
			if r.FormValue("gtin") != "7790895000997" {
				t.Errorf("gtin = %q", r.FormValue("gtin"))
			}
			_, _ = w.Write([]byte(successPage))
		default:
			http.NotFound(w, r)
		}
	}))

	got, err := client.Lookup(context.Background(), "7790895000997")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if got.Name != "COCA COLA GASEOSA Cola Regular" {
		t.Fatalf("name = %q", got.Name)
	}
	if got.Description == nil || *got.Description != "2,00 LT · PET · ACME" {
		t.Fatalf("description = %v", got.Description)
	}

	if _, err := client.Lookup(context.Background(), "7790895000997"); err != nil {
		t.Fatalf("second Lookup() error = %v", err)
	}
	if logins.Load() != 1 {
		t.Fatalf("logins = %d, want 1", logins.Load())
	}
}

func TestLookupDecodesHTMLAndSkipsEmptyFields(t *testing.T) {
	page := `
<p id="marca"><b>Marca: </b>TARAG&#220;I</p>
<p id="Descripcion"><b>Descripción: </b>TARAGÜI</p>
<p id="variedad"><b>Variedad: </b>-</p>
<p id="contenidoneto"><b>Contenido Neto: </b>1,00</p>
<p id="unidadmedida"><b>Unidad Medida: </b>KILOS</p>
`
	client := portal(t, signedIn(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(page))
	}))

	got, err := client.Lookup(context.Background(), "7790387013516")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if got.Name != "TARAGÜI" || got.Description == nil || *got.Description != "1,00 KILOS" {
		t.Fatalf("got = %+v", got)
	}
}

func TestLookupLimitsTheName(t *testing.T) {
	page := `<p id="Descripcion"><b>Descripción: </b>` + strings.Repeat("a", domain.MaxItemNameLength+25) + `</p>`
	client := portal(t, signedIn(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(page))
	}))

	got, err := client.Lookup(context.Background(), "1")
	if err != nil {
		t.Fatalf("Lookup() error = %v", err)
	}
	if len([]rune(got.Name)) != domain.MaxItemNameLength {
		t.Fatalf("name length = %d", len([]rune(got.Name)))
	}
}

func TestLookupReturnsNotFound(t *testing.T) {
	pages := []string{
		`<div id="divnoencontrado" class="alert alert-danger">no está registrado</div>`,
		`<div class="alert alert-success"></div><form id="frmGtin"></form>`,
	}
	for _, page := range pages {
		client := portal(t, signedIn(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(page))
		}))

		_, err := client.Lookup(context.Background(), "7790000000010")
		if !errors.Is(err, domain.ErrExternalProductNotFound) {
			t.Fatalf("Lookup() error = %v, want ErrExternalProductNotFound", err)
		}
	}
}

func TestLookupSignsInAgainWhenTheSessionIsRejected(t *testing.T) {
	var logins, lookups atomic.Int32
	client := portal(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case loginPath:
			logins.Add(1)
			http.SetCookie(w, &http.Cookie{Name: "session", Value: "ok", Path: "/"})
			_, _ = w.Write([]byte(`{"Status":"ok"}`))
		case lookupPath:
			if lookups.Add(1) == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte(successPage))
		default:
			http.NotFound(w, r)
		}
	}))

	got, err := client.Lookup(context.Background(), "7790895000997")
	if err != nil || got.Name == "" {
		t.Fatalf("Lookup() = %+v, %v", got, err)
	}
	if logins.Load() != 2 {
		t.Fatalf("logins = %d, want 2", logins.Load())
	}
}

func TestLookupFailsWhenTheSessionKeepsBeingRejected(t *testing.T) {
	client := portal(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == loginPath {
			_, _ = w.Write([]byte(`{"Status":"OK"}`))
			return
		}
		_, _ = w.Write([]byte(`<form id="frmGtin"></form>`))
	}))

	_, err := client.Lookup(context.Background(), "1")
	if err == nil || errors.Is(err, domain.ErrExternalProductNotFound) || !strings.Contains(err.Error(), "session") {
		t.Fatalf("Lookup() error = %v", err)
	}
}

func TestLookupFailsWhenLoginIsRejected(t *testing.T) {
	const secret = "s3cret-value"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"Status":"Fail"}`))
	}))
	t.Cleanup(server.Close)
	client := NewClient(server.Client(), server.URL, "member", secret, nil)

	_, err := client.Lookup(context.Background(), "1")
	if err == nil || errors.Is(err, domain.ErrExternalProductNotFound) || strings.Contains(err.Error(), secret) {
		t.Fatalf("Lookup() error = %v", err)
	}
}

func TestLookupFailsOnABadLoginResponse(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{"status", http.StatusInternalServerError, ""},
		{"not json", http.StatusOK, "nope"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := portal(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			}))

			_, err := client.Lookup(context.Background(), "1")
			if err == nil || errors.Is(err, domain.ErrExternalProductNotFound) {
				t.Fatalf("Lookup() error = %v", err)
			}
		})
	}
}

func TestLookupFailsOnAnUnexpectedPortalPage(t *testing.T) {
	client := portal(t, signedIn(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html>maintenance</html>"))
	}))

	_, err := client.Lookup(context.Background(), "1")
	if err == nil || errors.Is(err, domain.ErrExternalProductNotFound) {
		t.Fatalf("Lookup() error = %v", err)
	}
}

func TestLookupFailsWhenThePortalStatusIsUnexpected(t *testing.T) {
	client := portal(t, signedIn(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))

	_, err := client.Lookup(context.Background(), "1")
	if err == nil || errors.Is(err, domain.ErrExternalProductNotFound) {
		t.Fatalf("Lookup() error = %v", err)
	}
}

func TestLookupUsesTheNextCatalog(t *testing.T) {
	product := domain.ExternalProduct{Name: "From next"}
	next := &stubCatalog{product: product}

	t.Run("portal miss", func(t *testing.T) {
		client := portal(t, signedIn(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`<div id="divnoencontrado"></div>`))
		}))
		client.next = next

		got, err := client.Lookup(context.Background(), "1")
		if err != nil || got != product {
			t.Fatalf("Lookup() = %+v, %v", got, err)
		}
	})

	t.Run("portal down", func(t *testing.T) {
		client := NewClient(http.DefaultClient, "http://127.0.0.1:0", "member", "secret", next)

		got, err := client.Lookup(context.Background(), "1")
		if err != nil || got != product {
			t.Fatalf("Lookup() = %+v, %v", got, err)
		}
	})

	t.Run("no credentials", func(t *testing.T) {
		client := NewClient(nil, "http://127.0.0.1:0", "", "", next)

		got, err := client.Lookup(context.Background(), "1")
		if err != nil || got != product {
			t.Fatalf("Lookup() = %+v, %v", got, err)
		}
	})

	t.Run("does not ask next when the portal has the product", func(t *testing.T) {
		called := false
		client := portal(t, signedIn(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`<p id="marca"><b>Marca: </b>ACME</p>`))
		}))
		client.next = &stubCatalog{err: errors.New("next should not be called"), called: &called}

		got, err := client.Lookup(context.Background(), "1")
		if err != nil || got.Name != "ACME" || called {
			t.Fatalf("Lookup() = %+v, %v, called = %v", got, err, called)
		}
	})
}

func TestLookupReturnsNotFoundWhenEveryCatalogMisses(t *testing.T) {
	client := portal(t, signedIn(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<div id="divnoencontrado"></div>`))
	}))
	client.next = &stubCatalog{err: domain.ErrExternalProductNotFound}

	_, err := client.Lookup(context.Background(), "1")
	if !errors.Is(err, domain.ErrExternalProductNotFound) {
		t.Fatalf("Lookup() error = %v", err)
	}
}

func TestLookupWithoutCredentialsOrNextIsAMiss(t *testing.T) {
	client := NewClient(nil, "", "", "", nil)

	_, err := client.Lookup(context.Background(), "1")
	if !errors.Is(err, domain.ErrExternalProductNotFound) {
		t.Fatalf("Lookup() error = %v", err)
	}
}

func TestLookupPrefersThePortalErrorWhenNextAlsoFails(t *testing.T) {
	client := portal(t, signedIn(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	client.next = &stubCatalog{err: errors.New("next down")}

	_, err := client.Lookup(context.Background(), "1")
	if err == nil || !strings.Contains(err.Error(), "unexpected status") {
		t.Fatalf("Lookup() error = %v", err)
	}
}

func TestLookupReturnsNotFoundWhenThePortalFailsAndNextMisses(t *testing.T) {
	client := portal(t, signedIn(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	client.next = &stubCatalog{err: domain.ErrExternalProductNotFound}

	_, err := client.Lookup(context.Background(), "1")
	if !errors.Is(err, domain.ErrExternalProductNotFound) {
		t.Fatalf("Lookup() error = %v", err)
	}
}

func TestPackageNameKeepsACodeWithoutALabel(t *testing.T) {
	if got := packageName("(C7)"); got != "C7" {
		t.Fatalf("packageName = %q", got)
	}
	if code, text := splitCode("("); code != "" || text != "(" {
		t.Fatalf("splitCode = %q %q", code, text)
	}
}

func signedIn(lookup http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == loginPath {
			_, _ = io.WriteString(w, `{"Status":"OK"}`)
			return
		}
		lookup(w, r)
	})
}

type stubCatalog struct {
	product domain.ExternalProduct
	err     error
	called  *bool
}

func (stub *stubCatalog) Lookup(context.Context, string) (domain.ExternalProduct, error) {
	if stub.called != nil {
		*stub.called = true
	}
	if stub.err != nil {
		return domain.ExternalProduct{}, stub.err
	}

	return stub.product, nil
}

// Keep the compile-time check that the stub matches the contract used as next.
var _ gateway.ProductCatalog = (*stubCatalog)(nil)
