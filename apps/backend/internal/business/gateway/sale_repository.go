package gateway

import (
	"context"

	"market/backend/internal/business/domain"
)

// SaleFilter is a validated domain.SaleFilter plus the page to list.
type SaleFilter struct {
	domain.SaleFilter
	Page domain.PageRequest
}

// SaleRepository stores sales. Every implementation must pass the contract
// in infrastructure/repository/repositorytest.
//
// Sales are never deleted. A use case saves a sale, or its void, and records
// its stock movements within the same TxManager unit of work.
type SaleRepository interface {
	// Create assigns the ID and the next correlative Number. It returns
	// ErrIdempotencyKeyConflict when another sale has the same key,
	// ErrSaleCustomerNotFound or ErrSaleSellerNotFound when the customer or
	// the seller doesn't exist and ErrSaleItemNotFound when an item of a
	// line doesn't.
	Create(ctx context.Context, sale domain.Sale) (domain.Sale, error)
	// Get returns ErrSaleNotFound for an unknown ID.
	Get(ctx context.Context, id string) (domain.Sale, error)
	// GetByIdempotencyKey returns the sale made with key, or
	// ErrSaleNotFound when there is none.
	GetByIdempotencyKey(ctx context.Context, key string) (domain.Sale, error)
	// Void stores the status and the void fields of a sale the domain
	// voided; nothing else of the sale changes. It returns ErrSaleNotFound
	// for an unknown ID.
	Void(ctx context.Context, sale domain.Sale) (domain.Sale, error)
	// List returns one page of the sales that match the filter, newest
	// (highest number) first.
	List(ctx context.Context, filter SaleFilter) (domain.Page[domain.Sale], error)
}
