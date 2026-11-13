package gateway

import (
	"context"

	"market/backend/internal/business/domain"
)

type CategoryFilter struct {
	// Search matches the name or description, ignoring case and accents.
	Search string
	Page   domain.PageRequest
}

// CategoryRepository stores categories. Every implementation must pass the
// contract in infrastructure/repository/repositorytest.
type CategoryRepository interface {
	// Create assigns the ID. It returns ErrCategoryNameInUse when another
	// category already has the same NameKey.
	Create(ctx context.Context, category domain.Category) (domain.Category, error)
	// Get returns ErrCategoryNotFound for an unknown ID.
	Get(ctx context.Context, id string) (domain.Category, error)
	// Update replaces a stored category. It returns ErrCategoryNotFound for an
	// unknown ID and ErrCategoryNameInUse when the result would duplicate the
	// name of another category.
	Update(ctx context.Context, category domain.Category) (domain.Category, error)
	// Delete removes a category for good. It returns ErrCategoryNotFound for
	// an unknown ID and ErrCategoryInUse while items belong to it.
	Delete(ctx context.Context, id string) error
	// List returns one page sorted by name.
	List(ctx context.Context, filter CategoryFilter) (domain.Page[domain.Category], error)
}
