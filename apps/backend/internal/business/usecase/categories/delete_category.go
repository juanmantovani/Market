package categories

import (
	"context"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type DeleteCategoryInput struct {
	Actor domain.Actor
	ID    string
}

// DeleteCategory removes a category for good.
type DeleteCategory interface {
	Execute(ctx context.Context, input DeleteCategoryInput) error
}

type deleteCategory struct {
	repository gateway.CategoryRepository
}

func NewDeleteCategory(repository gateway.CategoryRepository) DeleteCategory {
	return &deleteCategory{repository: repository}
}

func (useCase *deleteCategory) Execute(ctx context.Context, input DeleteCategoryInput) error {
	if err := input.Actor.Authorize(domain.PermissionManageCatalog); err != nil {
		return err
	}

	return domain.FromStorage(useCase.repository.Delete(ctx, input.ID))
}
