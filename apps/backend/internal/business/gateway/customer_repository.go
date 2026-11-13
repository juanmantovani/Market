package gateway

import (
	"context"

	"market/backend/internal/business/domain"
)

type CustomerFilter struct {
	// Search matches the name, document number, email or phone, ignoring
	// case, accents and document separators.
	Search string
	Page   domain.PageRequest
}

// CustomerRepository stores customers. Every implementation must pass the
// contract in infrastructure/repository/repositorytest, and starts with the
// system customer domain.FinalConsumer.
type CustomerRepository interface {
	// Create assigns the ID. It returns ErrCustomerDocumentInUse when
	// another customer has the same DocumentKey.
	Create(ctx context.Context, customer domain.Customer) (domain.Customer, error)
	// Get returns ErrCustomerNotFound for an unknown ID.
	Get(ctx context.Context, id string) (domain.Customer, error)
	// Update replaces a stored customer, with the same errors as Create plus
	// ErrCustomerNotFound for an unknown ID.
	Update(ctx context.Context, customer domain.Customer) (domain.Customer, error)
	// Delete removes a customer for good. It returns ErrCustomerNotFound for
	// an unknown ID and ErrCustomerInUse while a sale names the customer.
	Delete(ctx context.Context, id string) error
	// List returns one page sorted by name.
	List(ctx context.Context, filter CustomerFilter) (domain.Page[domain.Customer], error)
}
