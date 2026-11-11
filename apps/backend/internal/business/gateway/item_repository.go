package gateway

import (
	"context"

	"market/backend/internal/business/domain"
)

type ItemFilter struct {
	// Search matches the name, code or barcode, ignoring case and accents.
	Search string
	// CategoryID, when set, keeps only the items of that category.
	CategoryID string
	Page       domain.PageRequest
}

// ItemRepository stores items. Every implementation must pass the contract
// in infrastructure/repository/repositorytest.
type ItemRepository interface {
	// Create assigns the ID. It returns ErrItemCodeInUse or
	// ErrItemBarcodeInUse for a duplicate, and ErrItemCategoryNotFound when
	// the category doesn't exist.
	Create(ctx context.Context, item domain.Item) (domain.Item, error)
	// Get returns ErrItemNotFound for an unknown ID.
	Get(ctx context.Context, id string) (domain.Item, error)
	// Update replaces a stored item, with the same errors as Create plus
	// ErrItemNotFound for an unknown ID.
	Update(ctx context.Context, item domain.Item) (domain.Item, error)
	// Delete removes an item for good. It returns ErrItemNotFound for an
	// unknown ID.
	Delete(ctx context.Context, id string) error
	// List returns one page sorted by name.
	List(ctx context.Context, filter ItemFilter) (domain.Page[domain.Item], error)
}
