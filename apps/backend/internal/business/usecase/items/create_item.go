package items

import (
	"context"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type CreateItemInput struct {
	Actor  domain.Actor
	Fields domain.ItemFields
}

type CreateItem interface {
	Execute(ctx context.Context, input CreateItemInput) (domain.Item, error)
}

type createItem struct {
	repository gateway.ItemRepository
}

func NewCreateItem(repository gateway.ItemRepository) CreateItem {
	return &createItem{repository: repository}
}

func (useCase *createItem) Execute(ctx context.Context, input CreateItemInput) (domain.Item, error) {
	if err := input.Actor.Authorize(domain.PermissionManageCatalog); err != nil {
		return domain.Item{}, err
	}

	item, err := domain.NewItem(input.Fields, time.Now().UTC())
	if err != nil {
		return domain.Item{}, err
	}

	created, err := useCase.repository.Create(ctx, item)
	if err != nil {
		return domain.Item{}, domain.FromStorage(err)
	}

	return created, nil
}
