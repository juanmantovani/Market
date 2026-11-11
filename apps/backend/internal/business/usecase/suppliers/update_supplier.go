package suppliers

import (
	"context"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type UpdateSupplierInput struct {
	Actor  domain.Actor
	ID     string
	Fields domain.SupplierFields
}

// UpdateSupplier replaces every editable field of a supplier.
type UpdateSupplier interface {
	Execute(ctx context.Context, input UpdateSupplierInput) (domain.Supplier, error)
}

type updateSupplier struct {
	repository gateway.SupplierRepository
}

func NewUpdateSupplier(repository gateway.SupplierRepository) UpdateSupplier {
	return &updateSupplier{repository: repository}
}

func (useCase *updateSupplier) Execute(ctx context.Context, input UpdateSupplierInput) (domain.Supplier, error) {
	if err := input.Actor.Authorize(domain.PermissionManageSuppliers); err != nil {
		return domain.Supplier{}, err
	}

	current, err := useCase.repository.Get(ctx, input.ID)
	if err != nil {
		return domain.Supplier{}, domain.FromStorage(err)
	}

	replaced, err := current.Replace(input.Fields, time.Now().UTC())
	if err != nil {
		return domain.Supplier{}, err
	}

	updated, err := useCase.repository.Update(ctx, replaced)
	if err != nil {
		return domain.Supplier{}, domain.FromStorage(err)
	}

	return updated, nil
}
