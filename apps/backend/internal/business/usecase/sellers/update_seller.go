package sellers

import (
	"context"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type UpdateSellerInput struct {
	Actor  domain.Actor
	ID     string
	Fields domain.SellerFields
}

// UpdateSeller replaces every editable field of a seller.
type UpdateSeller interface {
	Execute(ctx context.Context, input UpdateSellerInput) (domain.Seller, error)
}

type updateSeller struct {
	repository gateway.SellerRepository
}

func NewUpdateSeller(repository gateway.SellerRepository) UpdateSeller {
	return &updateSeller{repository: repository}
}

func (useCase *updateSeller) Execute(ctx context.Context, input UpdateSellerInput) (domain.Seller, error) {
	if err := input.Actor.Authorize(domain.PermissionManageSellers); err != nil {
		return domain.Seller{}, err
	}

	current, err := useCase.repository.Get(ctx, input.ID)
	if err != nil {
		return domain.Seller{}, domain.FromStorage(err)
	}

	replaced, err := current.Replace(input.Fields, time.Now().UTC())
	if err != nil {
		return domain.Seller{}, err
	}

	updated, err := useCase.repository.Update(ctx, replaced)
	if err != nil {
		return domain.Seller{}, domain.FromStorage(err)
	}

	return updated, nil
}
