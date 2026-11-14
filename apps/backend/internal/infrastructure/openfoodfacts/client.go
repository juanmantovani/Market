// Package openfoodfacts implements gateway.ProductCatalog against Open Food
// Facts' public read API (https://openfoodfacts.github.io/openfoodfacts-server/api/).
package openfoodfacts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

// BaseURL is Open Food Facts' production read API.
const BaseURL = "https://world.openfoodfacts.org"

// userAgent follows Open Food Facts' policy of identifying every API
// client: https://openfoodfacts.github.io/openfoodfacts-server/api/.
const userAgent = "MarketDemo/1.0 (https://github.com/market-demo)"

var _ gateway.ProductCatalog = (*Client)(nil)

// Client looks up a barcode in Open Food Facts' public catalog.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient uses http.DefaultClient and BaseURL when httpClient or baseURL
// are left empty.
func NewClient(httpClient *http.Client, baseURL string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if baseURL == "" {
		baseURL = BaseURL
	}

	return &Client{httpClient: httpClient, baseURL: baseURL}
}

type productResponse struct {
	Status  int `json:"status"`
	Product struct {
		Name     string `json:"product_name"`
		NameES   string `json:"product_name_es"`
		Brands   string `json:"brands"`
		Quantity string `json:"quantity"`
	} `json:"product"`
}

// Lookup calls GET /api/v2/product/{barcode}.json. It returns
// domain.ErrExternalProductNotFound both when Open Food Facts has nothing
// for barcode and when the request itself fails: a flaky third party
// shouldn't turn into a hard error for whoever is creating an item.
func (client *Client) Lookup(ctx context.Context, barcode string) (domain.ExternalProduct, error) {
	endpoint := fmt.Sprintf(
		"%s/api/v2/product/%s.json?fields=product_name,product_name_es,brands,quantity",
		client.baseURL, url.PathEscape(barcode),
	)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return domain.ExternalProduct{}, fmt.Errorf("openfoodfacts: building request: %w", err)
	}
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("Accept", "application/json")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return domain.ExternalProduct{}, fmt.Errorf("openfoodfacts: calling %s: %w", client.baseURL, err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return domain.ExternalProduct{}, domain.ErrExternalProductNotFound
	}
	if response.StatusCode != http.StatusOK {
		return domain.ExternalProduct{}, fmt.Errorf("openfoodfacts: unexpected status %d", response.StatusCode)
	}

	var decoded productResponse
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		return domain.ExternalProduct{}, fmt.Errorf("openfoodfacts: decoding response: %w", err)
	}
	if decoded.Status != 1 {
		return domain.ExternalProduct{}, domain.ErrExternalProductNotFound
	}

	name := strings.TrimSpace(decoded.Product.NameES)
	if name == "" {
		name = strings.TrimSpace(decoded.Product.Name)
	}
	if name == "" {
		return domain.ExternalProduct{}, domain.ErrExternalProductNotFound
	}

	return domain.ExternalProduct{
		Name:        name,
		Description: describe(decoded.Product.Brands, decoded.Product.Quantity),
	}, nil
}

// describe combines the brand and the package size into a short, optional
// description, e.g. "Ferrero, Nutella · 400 g".
func describe(brands, quantity string) *string {
	brands, quantity = strings.TrimSpace(brands), strings.TrimSpace(quantity)

	var text string
	switch {
	case brands != "" && quantity != "":
		text = brands + " · " + quantity
	case brands != "":
		text = brands
	case quantity != "":
		text = quantity
	default:
		return nil
	}

	return &text
}
