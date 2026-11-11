package items

import (
	"context"
	"strings"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type ListItemsInput struct {
	Actor      domain.Actor
	Search     string
	CategoryID string
	Page       int
	Limit      int
}

type ListItems interface {
	Execute(ctx context.Context, input ListItemsInput) (domain.Page[domain.Item], error)
}

type listItems struct {
	repository gateway.ItemRepository
}

func NewListItems(repository gateway.ItemRepository) ListItems {
	return &listItems{repository: repository}
}

func (useCase *listItems) Execute(ctx context.Context, input ListItemsInput) (domain.Page[domain.Item], error) {
	if err := input.Actor.Authorize(domain.PermissionViewCatalog); err != nil {
		return domain.Page[domain.Item]{}, err
	}

	page, err := domain.NewPageRequest(input.Page, input.Limit)
	if err != nil {
		return domain.Page[domain.Item]{}, err
	}

	result, err := useCase.repository.List(ctx, gateway.ItemFilter{
		Search:     strings.TrimSpace(input.Search),
		CategoryID: strings.TrimSpace(input.CategoryID),
		Page:       page,
	})
	if err != nil {
		return domain.Page[domain.Item]{}, domain.FromStorage(err)
	}

	return result, nil
}
