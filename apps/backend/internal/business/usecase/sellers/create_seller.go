package sellers

import (
	"context"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type CreateSellerInput struct {
	Actor  domain.Actor
	Fields domain.SellerFields
}

type CreateSeller interface {
	Execute(ctx context.Context, input CreateSellerInput) (domain.Seller, error)
}

type createSeller struct {
	repository gateway.SellerRepository
}

func NewCreateSeller(repository gateway.SellerRepository) CreateSeller {
	return &createSeller{repository: repository}
}

func (useCase *createSeller) Execute(ctx context.Context, input CreateSellerInput) (domain.Seller, error) {
	if err := input.Actor.Authorize(domain.PermissionManageSellers); err != nil {
		return domain.Seller{}, err
	}

	seller, err := domain.NewSeller(input.Fields, time.Now().UTC())
	if err != nil {
		return domain.Seller{}, err
	}

	created, err := useCase.repository.Create(ctx, seller)
	if err != nil {
		return domain.Seller{}, domain.FromStorage(err)
	}

	return created, nil
}
