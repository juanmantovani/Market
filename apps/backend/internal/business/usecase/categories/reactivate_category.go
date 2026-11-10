package categories

import (
	"context"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type ReactivateCategoryInput struct {
	Actor domain.Actor
	ID    string
}

// ReactivateCategory fails with ErrCategoryNameInUse when another active
// category took the name in the meantime.
type ReactivateCategory interface {
	Execute(ctx context.Context, input ReactivateCategoryInput) (domain.Category, error)
}

type reactivateCategory struct {
	repository gateway.CategoryRepository
	clock      gateway.Clock
}

func NewReactivateCategory(repository gateway.CategoryRepository, clock gateway.Clock) ReactivateCategory {
	return &reactivateCategory{repository: repository, clock: clock}
}

func (useCase *reactivateCategory) Execute(ctx context.Context, input ReactivateCategoryInput) (domain.Category, error) {
	return setActive(ctx, useCase.repository, useCase.clock, input.Actor, input.ID, true)
}
