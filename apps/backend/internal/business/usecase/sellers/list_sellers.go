package sellers

import (
	"context"
	"strings"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type ListSellersInput struct {
	Actor  domain.Actor
	Search string
	Page   int
	Limit  int
}

type ListSellers interface {
	Execute(ctx context.Context, input ListSellersInput) (domain.Page[domain.Seller], error)
}

type listSellers struct {
	repository gateway.SellerRepository
}

func NewListSellers(repository gateway.SellerRepository) ListSellers {
	return &listSellers{repository: repository}
}

func (useCase *listSellers) Execute(ctx context.Context, input ListSellersInput) (domain.Page[domain.Seller], error) {
	if err := input.Actor.Authorize(domain.PermissionViewSellers); err != nil {
		return domain.Page[domain.Seller]{}, err
	}

	page, err := domain.NewPageRequest(input.Page, input.Limit)
	if err != nil {
		return domain.Page[domain.Seller]{}, err
	}

	result, err := useCase.repository.List(ctx, gateway.SellerFilter{
		Search: strings.TrimSpace(input.Search),
		Page:   page,
	})
	if err != nil {
		return domain.Page[domain.Seller]{}, domain.FromStorage(err)
	}

	return result, nil
}
