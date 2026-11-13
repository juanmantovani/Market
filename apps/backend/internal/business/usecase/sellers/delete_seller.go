package sellers

import (
	"context"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type DeleteSellerInput struct {
	Actor domain.Actor
	ID    string
}

// DeleteSeller removes a seller for good. A seller with sales stays.
type DeleteSeller interface {
	Execute(ctx context.Context, input DeleteSellerInput) error
}

type deleteSeller struct {
	repository gateway.SellerRepository
}

func NewDeleteSeller(repository gateway.SellerRepository) DeleteSeller {
	return &deleteSeller{repository: repository}
}

func (useCase *deleteSeller) Execute(ctx context.Context, input DeleteSellerInput) error {
	if err := input.Actor.Authorize(domain.PermissionManageSellers); err != nil {
		return err
	}

	return domain.FromStorage(useCase.repository.Delete(ctx, input.ID))
}
