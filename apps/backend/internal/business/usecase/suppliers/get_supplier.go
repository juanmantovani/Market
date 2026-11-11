package suppliers

import (
	"context"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type GetSupplierInput struct {
	Actor domain.Actor
	ID    string
}

type GetSupplier interface {
	Execute(ctx context.Context, input GetSupplierInput) (domain.Supplier, error)
}

type getSupplier struct {
	repository gateway.SupplierRepository
}

func NewGetSupplier(repository gateway.SupplierRepository) GetSupplier {
	return &getSupplier{repository: repository}
}

func (useCase *getSupplier) Execute(ctx context.Context, input GetSupplierInput) (domain.Supplier, error) {
	if err := input.Actor.Authorize(domain.PermissionViewSuppliers); err != nil {
		return domain.Supplier{}, err
	}

	supplier, err := useCase.repository.Get(ctx, input.ID)
	if err != nil {
		return domain.Supplier{}, domain.FromStorage(err)
	}

	return supplier, nil
}
