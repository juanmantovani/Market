package customers

import (
	"context"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type GetCustomerInput struct {
	Actor domain.Actor
	ID    string
}

type GetCustomer interface {
	Execute(ctx context.Context, input GetCustomerInput) (domain.Customer, error)
}

type getCustomer struct {
	repository gateway.CustomerRepository
}

func NewGetCustomer(repository gateway.CustomerRepository) GetCustomer {
	return &getCustomer{repository: repository}
}

func (useCase *getCustomer) Execute(ctx context.Context, input GetCustomerInput) (domain.Customer, error) {
	if err := input.Actor.Authorize(domain.PermissionViewCustomers); err != nil {
		return domain.Customer{}, err
	}

	customer, err := useCase.repository.Get(ctx, input.ID)
	if err != nil {
		return domain.Customer{}, domain.FromStorage(err)
	}

	return customer, nil
}
