package items

import (
	"context"
	"strings"
	"unicode/utf8"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type LookupItemByBarcodeInput struct {
	Actor   domain.Actor
	Barcode string
}

// BarcodeLookup is what LookupItemByBarcode found for a barcode: Item is set
// when it's already in our catalog, Suggestion when it isn't but the
// external catalog has it. Neither is set when the barcode is unknown
// everywhere.
type BarcodeLookup struct {
	Item       *domain.Item
	Suggestion *domain.ExternalProduct
}

type LookupItemByBarcode interface {
	Execute(ctx context.Context, input LookupItemByBarcodeInput) (BarcodeLookup, error)
}

type lookupItemByBarcode struct {
	repository gateway.ItemRepository
	catalog    gateway.ProductCatalog
}

func NewLookupItemByBarcode(repository gateway.ItemRepository, catalog gateway.ProductCatalog) LookupItemByBarcode {
	return &lookupItemByBarcode{repository: repository, catalog: catalog}
}

func (useCase *lookupItemByBarcode) Execute(ctx context.Context, input LookupItemByBarcodeInput) (BarcodeLookup, error) {
	if err := input.Actor.Authorize(domain.PermissionManageCatalog); err != nil {
		return BarcodeLookup{}, err
	}

	barcode := strings.TrimSpace(input.Barcode)
	if barcode == "" {
		return BarcodeLookup{}, domain.ErrItemBarcodeRequired
	}
	if utf8.RuneCountInString(barcode) > domain.MaxItemBarcodeLength {
		return BarcodeLookup{}, domain.ErrItemBarcodeInvalid
	}

	found, err := useCase.findByBarcode(ctx, barcode)
	if err != nil {
		return BarcodeLookup{}, domain.FromStorage(err)
	}
	if found != nil {
		return BarcodeLookup{Item: found}, nil
	}

	// A miss in the external catalog, or any failure reaching it, just
	// means no suggestion: the person fills the item in by hand.
	product, err := useCase.catalog.Lookup(ctx, barcode)
	if err != nil {
		return BarcodeLookup{}, nil
	}

	return BarcodeLookup{Suggestion: &product}, nil
}

func (useCase *lookupItemByBarcode) findByBarcode(ctx context.Context, barcode string) (*domain.Item, error) {
	page, err := useCase.repository.List(ctx, gateway.ItemFilter{
		Search: barcode,
		Page:   domain.PageRequest{Page: 1, Limit: domain.MaxPageLimit},
	})
	if err != nil {
		return nil, err
	}
	for _, item := range page.Items {
		if item.Barcode != nil && *item.Barcode == barcode {
			return &item, nil
		}
	}

	return nil, nil
}
