package gateway

import (
	"context"

	"market/backend/internal/business/domain"
)

type SellerFilter struct {
	// Search matches the name or the code, ignoring case and accents.
	Search string
	Page   domain.PageRequest
}

// SellerRepository stores sellers. Every implementation must pass the
// contract in infrastructure/repository/repositorytest.
type SellerRepository interface {
	// Create assigns the ID. It returns ErrSellerCodeInUse when another
	// seller has the same code.
	Create(ctx context.Context, seller domain.Seller) (domain.Seller, error)
	// Get returns ErrSellerNotFound for an unknown ID.
	Get(ctx context.Context, id string) (domain.Seller, error)
	// Update replaces a stored seller, with the same errors as Create plus
	// ErrSellerNotFound for an unknown ID.
	Update(ctx context.Context, seller domain.Seller) (domain.Seller, error)
	// Delete removes a seller for good. It returns ErrSellerNotFound for
	// an unknown ID and ErrSellerInUse while a sale names the seller.
	Delete(ctx context.Context, id string) error
	// List returns one page sorted by name.
	List(ctx context.Context, filter SellerFilter) (domain.Page[domain.Seller], error)
}
