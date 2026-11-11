package items

import (
	"context"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type UpdateItemInput struct {
	Actor  domain.Actor
	ID     string
	Fields domain.ItemFields
}

// UpdateItem replaces every editable field of an item.
type UpdateItem interface {
	Execute(ctx context.Context, input UpdateItemInput) (domain.Item, error)
}

type updateItem struct {
	repository gateway.ItemRepository
}

func NewUpdateItem(repository gateway.ItemRepository) UpdateItem {
	return &updateItem{repository: repository}
}

func (useCase *updateItem) Execute(ctx context.Context, input UpdateItemInput) (domain.Item, error) {
	if err := input.Actor.Authorize(domain.PermissionManageCatalog); err != nil {
		return domain.Item{}, err
	}

	current, err := useCase.repository.Get(ctx, input.ID)
	if err != nil {
		return domain.Item{}, domain.FromStorage(err)
	}

	replaced, err := current.Replace(input.Fields, time.Now().UTC())
	if err != nil {
		return domain.Item{}, err
	}

	updated, err := useCase.repository.Update(ctx, replaced)
	if err != nil {
		return domain.Item{}, domain.FromStorage(err)
	}

	return updated, nil
}
