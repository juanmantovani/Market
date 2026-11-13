package customers

import (
	"context"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type CreateCustomerInput struct {
	Actor  domain.Actor
	Fields domain.CustomerFields
}

type CreateCustomer interface {
	Execute(ctx context.Context, input CreateCustomerInput) (domain.Customer, error)
}

type createCustomer struct {
	repository gateway.CustomerRepository
}

func NewCreateCustomer(repository gateway.CustomerRepository) CreateCustomer {
	return &createCustomer{repository: repository}
}

func (useCase *createCustomer) Execute(ctx context.Context, input CreateCustomerInput) (domain.Customer, error) {
	if err := input.Actor.Authorize(domain.PermissionCreateCustomers); err != nil {
		return domain.Customer{}, err
	}

	customer, err := domain.NewCustomer(input.Fields, time.Now().UTC())
	if err != nil {
		return domain.Customer{}, err
	}

	created, err := useCase.repository.Create(ctx, customer)
	if err != nil {
		return domain.Customer{}, domain.FromStorage(err)
	}

	return created, nil
}
