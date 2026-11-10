package categories

import (
	"context"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type UpdateCategoryInput struct {
	Actor   domain.Actor
	ID      string
	Changes domain.CategoryChanges
}

type UpdateCategory interface {
	Execute(ctx context.Context, input UpdateCategoryInput) (domain.Category, error)
}

type updateCategory struct {
	repository gateway.CategoryRepository
	clock      gateway.Clock
}

func NewUpdateCategory(repository gateway.CategoryRepository, clock gateway.Clock) UpdateCategory {
	return &updateCategory{repository: repository, clock: clock}
}

func (useCase *updateCategory) Execute(ctx context.Context, input UpdateCategoryInput) (domain.Category, error) {
	if err := input.Actor.Authorize(manageRoles...); err != nil {
		return domain.Category{}, err
	}
	if input.Changes.IsEmpty() {
		return domain.Category{}, domain.ErrCategoryUpdateEmpty
	}

	current, err := useCase.repository.Get(ctx, input.ID)
	if err != nil {
		return domain.Category{}, domain.FromStorage(err)
	}

	changed, err := current.Apply(input.Changes, useCase.clock.Now())
	if err != nil {
		return domain.Category{}, err
	}

	updated, err := useCase.repository.Update(ctx, changed)
	if err != nil {
		return domain.Category{}, domain.FromStorage(err)
	}

	return updated, nil
}
