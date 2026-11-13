package customers

import (
	"context"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type DeleteCustomerInput struct {
	Actor domain.Actor
	ID    string
}

// DeleteCustomer removes a customer for good. Consumidor Final and
// customers with sales stay.
type DeleteCustomer interface {
	Execute(ctx context.Context, input DeleteCustomerInput) error
}

type deleteCustomer struct {
	repository gateway.CustomerRepository
}

func NewDeleteCustomer(repository gateway.CustomerRepository) DeleteCustomer {
	return &deleteCustomer{repository: repository}
}

func (useCase *deleteCustomer) Execute(ctx context.Context, input DeleteCustomerInput) error {
	if err := input.Actor.Authorize(domain.PermissionManageCustomers); err != nil {
		return err
	}

	customer, err := useCase.repository.Get(ctx, input.ID)
	if err != nil {
		return domain.FromStorage(err)
	}
	if err := customer.EnsureChangeable(); err != nil {
		return err
	}

	return domain.FromStorage(useCase.repository.Delete(ctx, input.ID))
}
