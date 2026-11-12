package sales

import (
	"context"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type GetSaleInput struct {
	Actor domain.Actor
	ID    string
}

type GetSale interface {
	Execute(ctx context.Context, input GetSaleInput) (domain.Sale, error)
}

type getSale struct {
	repository gateway.SaleRepository
}

func NewGetSale(repository gateway.SaleRepository) GetSale {
	return &getSale{repository: repository}
}

func (useCase *getSale) Execute(ctx context.Context, input GetSaleInput) (domain.Sale, error) {
	if err := input.Actor.Authorize(domain.PermissionViewSales); err != nil {
		return domain.Sale{}, err
	}

	sale, err := useCase.repository.Get(ctx, input.ID)
	if err != nil {
		return domain.Sale{}, domain.FromStorage(err)
	}

	return sale, nil
}
