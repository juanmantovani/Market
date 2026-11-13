package sales

import (
	"context"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type ListSalesInput struct {
	Actor  domain.Actor
	Filter domain.SaleFilter
	Page   int
	Limit  int
}

// ListSales lists sales newest first, optionally only those of a period, a
// customer, a seller or a status.
type ListSales interface {
	Execute(ctx context.Context, input ListSalesInput) (domain.Page[domain.Sale], error)
}

type listSales struct {
	repository gateway.SaleRepository
}

func NewListSales(repository gateway.SaleRepository) ListSales {
	return &listSales{repository: repository}
}

func (useCase *listSales) Execute(ctx context.Context, input ListSalesInput) (domain.Page[domain.Sale], error) {
	if err := input.Actor.Authorize(domain.PermissionViewSales); err != nil {
		return domain.Page[domain.Sale]{}, err
	}

	filter, err := domain.NewSaleFilter(input.Filter)
	if err != nil {
		return domain.Page[domain.Sale]{}, err
	}
	page, err := domain.NewPageRequest(input.Page, input.Limit)
	if err != nil {
		return domain.Page[domain.Sale]{}, err
	}

	result, err := useCase.repository.List(ctx, gateway.SaleFilter{SaleFilter: filter, Page: page})
	if err != nil {
		return domain.Page[domain.Sale]{}, domain.FromStorage(err)
	}

	return result, nil
}
