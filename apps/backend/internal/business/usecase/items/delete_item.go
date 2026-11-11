package items

import (
	"context"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type DeleteItemInput struct {
	Actor domain.Actor
	ID    string
}

// DeleteItem removes an item for good.
type DeleteItem interface {
	Execute(ctx context.Context, input DeleteItemInput) error
}

type deleteItem struct {
	repository gateway.ItemRepository
}

func NewDeleteItem(repository gateway.ItemRepository) DeleteItem {
	return &deleteItem{repository: repository}
}

func (useCase *deleteItem) Execute(ctx context.Context, input DeleteItemInput) error {
	if err := input.Actor.Authorize(domain.PermissionManageCatalog); err != nil {
		return err
	}

	return domain.FromStorage(useCase.repository.Delete(ctx, input.ID))
}
