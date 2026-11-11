package items

import (
	"context"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type GetItemInput struct {
	Actor domain.Actor
	ID    string
}

type GetItem interface {
	Execute(ctx context.Context, input GetItemInput) (domain.Item, error)
}

type getItem struct {
	repository gateway.ItemRepository
}

func NewGetItem(repository gateway.ItemRepository) GetItem {
	return &getItem{repository: repository}
}

func (useCase *getItem) Execute(ctx context.Context, input GetItemInput) (domain.Item, error) {
	if err := input.Actor.Authorize(domain.PermissionViewCatalog); err != nil {
		return domain.Item{}, err
	}

	item, err := useCase.repository.Get(ctx, input.ID)
	if err != nil {
		return domain.Item{}, domain.FromStorage(err)
	}

	return item, nil
}
