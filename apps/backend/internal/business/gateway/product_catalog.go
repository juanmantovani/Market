package gateway

import (
	"context"

	"market/backend/internal/business/domain"
)

// ProductCatalog looks up public product data by barcode in an external
// catalog, to suggest values for a new item that isn't in our own catalog
// yet. Every implementation returns domain.ErrExternalProductNotFound when
// it has nothing for barcode.
type ProductCatalog interface {
	Lookup(ctx context.Context, barcode string) (domain.ExternalProduct, error)
}
