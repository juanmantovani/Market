package suppliers

import (
	"context"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type DeleteSupplierInput struct {
	Actor domain.Actor
	ID    string
}

// DeleteSupplier removes a supplier for good.
type DeleteSupplier interface {
	Execute(ctx context.Context, input DeleteSupplierInput) error
}

type deleteSupplier struct {
	repository gateway.SupplierRepository
}

func NewDeleteSupplier(repository gateway.SupplierRepository) DeleteSupplier {
	return &deleteSupplier{repository: repository}
}

func (useCase *deleteSupplier) Execute(ctx context.Context, input DeleteSupplierInput) error {
	if err := input.Actor.Authorize(domain.PermissionManageSuppliers); err != nil {
		return err
	}

	return domain.FromStorage(useCase.repository.Delete(ctx, input.ID))
}
