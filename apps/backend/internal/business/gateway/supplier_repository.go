package gateway

import (
	"context"

	"market/backend/internal/business/domain"
)

type SupplierFilter struct {
	// Search matches the business name, contact name, CUIT, email or phone,
	// ignoring case, accents and CUIT dashes.
	Search string
	Page   domain.PageRequest
}

// SupplierRepository stores suppliers. Every implementation must pass the
// contract in infrastructure/repository/repositorytest.
type SupplierRepository interface {
	// Create assigns the ID. It returns ErrSupplierCUITInUse when another
	// supplier has the same CUIT; suppliers without one never collide.
	Create(ctx context.Context, supplier domain.Supplier) (domain.Supplier, error)
	// Get returns ErrSupplierNotFound for an unknown ID.
	Get(ctx context.Context, id string) (domain.Supplier, error)
	// Update replaces a stored supplier, with the same errors as Create plus
	// ErrSupplierNotFound for an unknown ID.
	Update(ctx context.Context, supplier domain.Supplier) (domain.Supplier, error)
	// Delete removes a supplier for good. It returns ErrSupplierNotFound for
	// an unknown ID and ErrSupplierInUse while an item lists the supplier.
	Delete(ctx context.Context, id string) error
	// List returns one page sorted by business name.
	List(ctx context.Context, filter SupplierFilter) (domain.Page[domain.Supplier], error)
}
