package categories

import (
	"context"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type GetCategoryInput struct {
	Actor domain.Actor
	ID    string
}

type GetCategory interface {
	Execute(ctx context.Context, input GetCategoryInput) (domain.Category, error)
}

type getCategory struct {
	repository gateway.CategoryRepository
}

func NewGetCategory(repository gateway.CategoryRepository) GetCategory {
	return &getCategory{repository: repository}
}

func (useCase *getCategory) Execute(ctx context.Context, input GetCategoryInput) (domain.Category, error) {
	if err := input.Actor.Authorize(readRoles...); err != nil {
		return domain.Category{}, err
	}

	category, err := useCase.repository.Get(ctx, input.ID)
	if err != nil {
		return domain.Category{}, domain.FromStorage(err)
	}

	return category, nil
}
