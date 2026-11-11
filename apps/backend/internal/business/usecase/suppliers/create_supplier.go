package suppliers

import (
	"context"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type CreateSupplierInput struct {
	Actor  domain.Actor
	Fields domain.SupplierFields
}

type CreateSupplier interface {
	Execute(ctx context.Context, input CreateSupplierInput) (domain.Supplier, error)
}

type createSupplier struct {
	repository gateway.SupplierRepository
}

func NewCreateSupplier(repository gateway.SupplierRepository) CreateSupplier {
	return &createSupplier{repository: repository}
}

func (useCase *createSupplier) Execute(ctx context.Context, input CreateSupplierInput) (domain.Supplier, error) {
	if err := input.Actor.Authorize(domain.PermissionManageSuppliers); err != nil {
		return domain.Supplier{}, err
	}

	supplier, err := domain.NewSupplier(input.Fields, time.Now().UTC())
	if err != nil {
		return domain.Supplier{}, err
	}

	created, err := useCase.repository.Create(ctx, supplier)
	if err != nil {
		return domain.Supplier{}, domain.FromStorage(err)
	}

	return created, nil
}
