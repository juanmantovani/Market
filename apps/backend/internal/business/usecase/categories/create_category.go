package categories

import (
	"context"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type CreateCategoryInput struct {
	Actor       domain.Actor
	Name        string
	Description *string
}

type CreateCategory interface {
	Execute(ctx context.Context, input CreateCategoryInput) (domain.Category, error)
}

type createCategory struct {
	repository gateway.CategoryRepository
}

func NewCreateCategory(repository gateway.CategoryRepository) CreateCategory {
	return &createCategory{repository: repository}
}

func (useCase *createCategory) Execute(ctx context.Context, input CreateCategoryInput) (domain.Category, error) {
	if err := input.Actor.Authorize(domain.PermissionManageCatalog); err != nil {
		return domain.Category{}, err
	}

	category, err := domain.NewCategory(input.Name, input.Description, time.Now().UTC())
	if err != nil {
		return domain.Category{}, err
	}

	created, err := useCase.repository.Create(ctx, category)
	if err != nil {
		return domain.Category{}, domain.FromStorage(err)
	}

	return created, nil
}
