package suppliers

import (
	"context"
	"strings"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type ListSuppliersInput struct {
	Actor  domain.Actor
	Search string
	Page   int
	Limit  int
}

type ListSuppliers interface {
	Execute(ctx context.Context, input ListSuppliersInput) (domain.Page[domain.Supplier], error)
}

type listSuppliers struct {
	repository gateway.SupplierRepository
}

func NewListSuppliers(repository gateway.SupplierRepository) ListSuppliers {
	return &listSuppliers{repository: repository}
}

func (useCase *listSuppliers) Execute(ctx context.Context, input ListSuppliersInput) (domain.Page[domain.Supplier], error) {
	if err := input.Actor.Authorize(domain.PermissionViewSuppliers); err != nil {
		return domain.Page[domain.Supplier]{}, err
	}

	page, err := domain.NewPageRequest(input.Page, input.Limit)
	if err != nil {
		return domain.Page[domain.Supplier]{}, err
	}

	result, err := useCase.repository.List(ctx, gateway.SupplierFilter{
		Search: strings.TrimSpace(input.Search),
		Page:   page,
	})
	if err != nil {
		return domain.Page[domain.Supplier]{}, domain.FromStorage(err)
	}

	return result, nil
}
