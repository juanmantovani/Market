package gatewayfake

import (
	"context"
	"fmt"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

// CustomerRepository is a fake gateway.CustomerRepository for use case
// tests. It records every call and returns the configured results.
type CustomerRepository struct {
	CreateErr error
	Created   []domain.Customer

	GetResult domain.Customer
	GetErr    error
	GetIDs    []string

	UpdateErr error
	Updated   []domain.Customer

	DeleteErr  error
	DeletedIDs []string

	ListResult  domain.Page[domain.Customer]
	ListErr     error
	ListFilters []gateway.CustomerFilter
}

func (fake *CustomerRepository) Create(_ context.Context, customer domain.Customer) (domain.Customer, error) {
	if fake.CreateErr != nil {
		return domain.Customer{}, fake.CreateErr
	}
	customer.ID = fmt.Sprintf("customer-%d", len(fake.Created)+1)
	fake.Created = append(fake.Created, customer)

	return customer, nil
}

func (fake *CustomerRepository) Get(_ context.Context, id string) (domain.Customer, error) {
	fake.GetIDs = append(fake.GetIDs, id)
	if fake.GetErr != nil {
		return domain.Customer{}, fake.GetErr
	}

	return fake.GetResult, nil
}

func (fake *CustomerRepository) Update(_ context.Context, customer domain.Customer) (domain.Customer, error) {
	if fake.UpdateErr != nil {
		return domain.Customer{}, fake.UpdateErr
	}
	fake.Updated = append(fake.Updated, customer)

	return customer, nil
}

func (fake *CustomerRepository) Delete(_ context.Context, id string) error {
	if fake.DeleteErr != nil {
		return fake.DeleteErr
	}
	fake.DeletedIDs = append(fake.DeletedIDs, id)

	return nil
}

func (fake *CustomerRepository) List(_ context.Context, filter gateway.CustomerFilter) (domain.Page[domain.Customer], error) {
	fake.ListFilters = append(fake.ListFilters, filter)
	if fake.ListErr != nil {
		return domain.Page[domain.Customer]{}, fake.ListErr
	}

	return fake.ListResult, nil
}
