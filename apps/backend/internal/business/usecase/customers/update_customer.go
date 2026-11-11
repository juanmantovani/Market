package customers

import (
	"context"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type UpdateCustomerInput struct {
	Actor  domain.Actor
	ID     string
	Fields domain.CustomerFields
}

// UpdateCustomer replaces every editable field of a customer.
type UpdateCustomer interface {
	Execute(ctx context.Context, input UpdateCustomerInput) (domain.Customer, error)
}

type updateCustomer struct {
	repository gateway.CustomerRepository
}

func NewUpdateCustomer(repository gateway.CustomerRepository) UpdateCustomer {
	return &updateCustomer{repository: repository}
}

func (useCase *updateCustomer) Execute(ctx context.Context, input UpdateCustomerInput) (domain.Customer, error) {
	if err := input.Actor.Authorize(domain.PermissionManageCustomers); err != nil {
		return domain.Customer{}, err
	}

	current, err := useCase.repository.Get(ctx, input.ID)
	if err != nil {
		return domain.Customer{}, domain.FromStorage(err)
	}

	replaced, err := current.Replace(input.Fields, time.Now().UTC())
	if err != nil {
		return domain.Customer{}, err
	}

	updated, err := useCase.repository.Update(ctx, replaced)
	if err != nil {
		return domain.Customer{}, domain.FromStorage(err)
	}

	return updated, nil
}
