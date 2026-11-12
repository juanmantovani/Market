package gateway

import (
	"context"

	"market/backend/internal/business/domain"
)

// SaleRepository stores sales. Every implementation must pass the contract
// in infrastructure/repository/repositorytest.
//
// Sales are never deleted. A use case saves a sale and records its stock
// movements within the same TxManager unit of work.
type SaleRepository interface {
	// Create assigns the ID and the next correlative Number. It returns
	// ErrIdempotencyKeyConflict when another sale has the same key,
	// ErrSaleCustomerNotFound or ErrSaleSellerNotFound when the customer or
	// the seller doesn't exist and ErrSaleItemNotFound when an item of a
	// line doesn't.
	Create(ctx context.Context, sale domain.Sale) (domain.Sale, error)
	// GetByIdempotencyKey returns the sale made with key, or
	// ErrSaleNotFound when there is none.
	GetByIdempotencyKey(ctx context.Context, key string) (domain.Sale, error)
}
