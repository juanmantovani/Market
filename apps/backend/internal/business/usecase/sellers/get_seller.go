package sellers

import (
	"context"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type GetSellerInput struct {
	Actor domain.Actor
	ID    string
}

type GetSeller interface {
	Execute(ctx context.Context, input GetSellerInput) (domain.Seller, error)
}

type getSeller struct {
	repository gateway.SellerRepository
}

func NewGetSeller(repository gateway.SellerRepository) GetSeller {
	return &getSeller{repository: repository}
}

func (useCase *getSeller) Execute(ctx context.Context, input GetSellerInput) (domain.Seller, error) {
	if err := input.Actor.Authorize(domain.PermissionViewSellers); err != nil {
		return domain.Seller{}, err
	}

	seller, err := useCase.repository.Get(ctx, input.ID)
	if err != nil {
		return domain.Seller{}, domain.FromStorage(err)
	}

	return seller, nil
}
