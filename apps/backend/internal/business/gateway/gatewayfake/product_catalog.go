package gatewayfake

import (
	"context"

	"market/backend/internal/business/domain"
)

// ProductCatalog is a fake gateway.ProductCatalog for use case tests. It
// records every barcode it was asked for and returns the configured result.
type ProductCatalog struct {
	Result   domain.ExternalProduct
	Err      error
	Barcodes []string
}

func (fake *ProductCatalog) Lookup(_ context.Context, barcode string) (domain.ExternalProduct, error) {
	fake.Barcodes = append(fake.Barcodes, barcode)
	if fake.Err != nil {
		return domain.ExternalProduct{}, fake.Err
	}

	return fake.Result, nil
}
