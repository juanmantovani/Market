// Package gs1argentina implements gateway.ProductCatalog against the GS1
// Argentina member portal. Verified by GS1 has no public API yet; the portal
// the member already uses accepts one GTIN per POST to
// /Producto/ConsultaVerified. Replace this adapter when that API ships.
package gs1argentina

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

// BaseURL is the GS1 Argentina member portal.
const BaseURL = "https://www.gs1.org.ar"

const (
	loginPath  = "/Account/AjaxLogin/"
	lookupPath = "/Producto/ConsultaVerified"
	userAgent  = "Market/1.0 (GS1 Argentina member portal)"
)

// errSessionExpired means the portal answered as if nobody was signed in.
// Lookup signs in again and retries the GTIN once.
var errSessionExpired = errors.New("gs1: session expired")

var fieldPattern = regexp.MustCompile(`(?i)<p id="([^"]+)"[^>]*>\s*(?:<b>[^<]*</b>)?\s*([^<]*)`)

var _ gateway.ProductCatalog = (*Client)(nil)

// Client looks up one GTIN at a time with a member account. next, when set,
// is asked whenever the portal has no product or cannot be reached.
type Client struct {
	httpClient *http.Client
	baseURL    string
	username   string
	password   string
	next       gateway.ProductCatalog

	mu       sync.Mutex
	loggedIn bool
}

// NewClient uses a 15s timeout and BaseURL when httpClient or baseURL are
// left empty. username and password are the member portal credentials; both
// empty disables the portal and leaves only next.
func NewClient(httpClient *http.Client, baseURL, username, password string, next gateway.ProductCatalog) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	if baseURL == "" {
		baseURL = BaseURL
	}

	// The portal keeps the session in a cookie. The jar stays on our client
	// so a shared http.Client is not mutated.
	jar, _ := cookiejar.New(nil)

	return &Client{
		httpClient: &http.Client{
			Transport:     httpClient.Transport,
			CheckRedirect: httpClient.CheckRedirect,
			Timeout:       httpClient.Timeout,
			Jar:           jar,
		},
		baseURL:  strings.TrimRight(baseURL, "/"),
		username: username,
		password: password,
		next:     next,
	}
}

// Lookup signs in, posts the GTIN and reads the product fields the portal
// renders. A miss, and a portal that cannot be reached, fall through to next.
func (client *Client) Lookup(ctx context.Context, barcode string) (domain.ExternalProduct, error) {
	if client.username == "" || client.password == "" {
		if client.next == nil {
			return domain.ExternalProduct{}, domain.ErrExternalProductNotFound
		}

		return client.next.Lookup(ctx, barcode)
	}

	product, err := client.lookupPortal(ctx, barcode)
	if err == nil {
		return product, nil
	}

	return client.lookupNext(ctx, barcode, err)
}

func (client *Client) lookupNext(ctx context.Context, barcode string, portalErr error) (domain.ExternalProduct, error) {
	if client.next == nil {
		return domain.ExternalProduct{}, portalErr
	}

	product, err := client.next.Lookup(ctx, barcode)
	if err == nil {
		return product, nil
	}
	if errors.Is(portalErr, domain.ErrExternalProductNotFound) || errors.Is(err, domain.ErrExternalProductNotFound) {
		return domain.ExternalProduct{}, domain.ErrExternalProductNotFound
	}

	return domain.ExternalProduct{}, portalErr
}

func (client *Client) lookupPortal(ctx context.Context, barcode string) (domain.ExternalProduct, error) {
	if err := client.ensureLogin(ctx); err != nil {
		return domain.ExternalProduct{}, err
	}

	product, err := client.query(ctx, barcode)
	if !errors.Is(err, errSessionExpired) {
		return product, err
	}

	client.invalidate()
	if err := client.ensureLogin(ctx); err != nil {
		return domain.ExternalProduct{}, err
	}

	product, err = client.query(ctx, barcode)
	if errors.Is(err, errSessionExpired) {
		return domain.ExternalProduct{}, fmt.Errorf("gs1: portal session was rejected")
	}

	return product, err
}

func (client *Client) ensureLogin(ctx context.Context) error {
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.loggedIn {
		return nil
	}

	if err := client.login(ctx); err != nil {
		return err
	}
	client.loggedIn = true

	return nil
}

func (client *Client) invalidate() {
	client.mu.Lock()
	client.loggedIn = false
	client.mu.Unlock()
}

func (client *Client) login(ctx context.Context) error {
	form := url.Values{}
	form.Set("NombreUsuario", client.username)
	form.Set("Password", client.password)
	form.Set("ReturnUrl", "")

	response, err := client.post(ctx, loginPath, form)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("gs1: login status %d", response.StatusCode)
	}

	var payload struct {
		Status string `json:"Status"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload); err != nil {
		return fmt.Errorf("gs1: login response is not JSON: %w", err)
	}
	if !strings.EqualFold(payload.Status, "OK") {
		return fmt.Errorf("gs1: login rejected")
	}

	return nil
}

func (client *Client) query(ctx context.Context, barcode string) (domain.ExternalProduct, error) {
	form := url.Values{}
	form.Set("gtin", barcode)

	response, err := client.post(ctx, lookupPath, form)
	if err != nil {
		return domain.ExternalProduct{}, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return domain.ExternalProduct{}, errSessionExpired
	}
	if response.StatusCode != http.StatusOK {
		return domain.ExternalProduct{}, fmt.Errorf("gs1: unexpected status %d", response.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return domain.ExternalProduct{}, fmt.Errorf("gs1: reading response: %w", err)
	}

	return parseVerified(string(body))
}

func (client *Client) post(ctx context.Context, path string, form url.Values) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.baseURL+path, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("gs1: building request: %w", err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json, text/html")
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("X-Requested-With", "XMLHttpRequest")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("gs1: calling %s: %w", client.baseURL, err)
	}

	return response, nil
}

// parseVerified reads the product block the portal renders for one GTIN.
// divnoencontrado is the portal's own "not registered" alert.
func parseVerified(body string) (domain.ExternalProduct, error) {
	if strings.Contains(body, `id="divnoencontrado"`) {
		return domain.ExternalProduct{}, domain.ErrExternalProductNotFound
	}

	fields := map[string]string{}
	for _, match := range fieldPattern.FindAllStringSubmatch(body, -1) {
		fields[strings.ToLower(match[1])] = clean(match[2])
	}

	name := productName(fields["marca"], fields["descripcion"], fields["variedad"])
	if name != "" {
		description := productDescription(
			quantity(fields["contenidoneto"], fields["unidadmedida"]),
			packageName(fields["envaseg"]),
			fields["empresa"],
		)

		return domain.ExternalProduct{Name: name, Description: description}, nil
	}

	if strings.Contains(body, "alert-success") {
		return domain.ExternalProduct{}, domain.ErrExternalProductNotFound
	}
	if strings.Contains(body, `id="frmGtin"`) || strings.Contains(body, "user-login") || strings.Contains(body, "/Account/AjaxLogin") {
		return domain.ExternalProduct{}, errSessionExpired
	}

	return domain.ExternalProduct{}, fmt.Errorf("gs1: unexpected portal page")
}

func productName(brand, description, variety string) string {
	name := description
	if name == "" {
		name = brand
	} else if brand != "" && !containsFold(name, brand) {
		name = brand + " " + name
	}
	if variety != "" && !containsFold(name, variety) {
		name += " " + variety
	}

	return limitRunes(name, domain.MaxItemNameLength)
}

func productDescription(parts ...string) *string {
	var kept []string
	for _, part := range parts {
		if part != "" {
			kept = append(kept, part)
		}
	}
	if len(kept) == 0 {
		return nil
	}

	text := limitRunes(strings.Join(kept, " · "), domain.MaxItemDescriptionLength)

	return &text
}

func quantity(content, unit string) string {
	code, text := splitCode(unit)
	label := code
	if label == "" {
		label = text
	}

	switch {
	case content != "" && label != "":
		return content + " " + label
	case content != "":
		return content
	default:
		return label
	}
}

// packageName prefers the words after a "(code)" prefix: "(C7) PET" is PET.
func packageName(value string) string {
	code, text := splitCode(value)
	if text != "" {
		return text
	}

	return code
}

// splitCode splits "(LT) LITROS" into "LT" and "LITROS".
func splitCode(value string) (string, string) {
	if !strings.HasPrefix(value, "(") {
		return "", value
	}
	end := strings.Index(value, ")")
	if end <= 1 {
		return "", value
	}

	return strings.TrimSpace(value[1:end]), strings.TrimSpace(value[end+1:])
}

func clean(value string) string {
	value = strings.TrimSpace(html.UnescapeString(value))
	if value == "-" || value == "—" {
		return ""
	}

	return value
}

func containsFold(value, part string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(part))
}

func limitRunes(value string, max int) string {
	if utf8.RuneCountInString(value) <= max {
		return value
	}

	return string([]rune(value)[:max])
}
