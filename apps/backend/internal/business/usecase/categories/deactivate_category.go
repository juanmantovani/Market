package categories

import (
	"context"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type DeactivateCategoryInput struct {
	Actor domain.Actor
	ID    string
}

// DeactivateCategory is how a category is removed. It stays stored so past
// records keep pointing at it, and it can be reactivated.
type DeactivateCategory interface {
	Execute(ctx context.Context, input DeactivateCategoryInput) (domain.Category, error)
}

type deactivateCategory struct {
	repository gateway.CategoryRepository
	clock      gateway.Clock
}

func NewDeactivateCategory(repository gateway.CategoryRepository, clock gateway.Clock) DeactivateCategory {
	return &deactivateCategory{repository: repository, clock: clock}
}

func (useCase *deactivateCategory) Execute(ctx context.Context, input DeactivateCategoryInput) (domain.Category, error) {
	return setActive(ctx, useCase.repository, useCase.clock, input.Actor, input.ID, false)
}
