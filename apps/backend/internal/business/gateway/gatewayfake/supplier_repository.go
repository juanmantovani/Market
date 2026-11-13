package gatewayfake

import (
	"context"
	"fmt"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

// SupplierRepository is a fake gateway.SupplierRepository for use case
// tests. It records every call and returns the configured results.
type SupplierRepository struct {
	CreateErr error
	Created   []domain.Supplier

	GetResult domain.Supplier
	GetErr    error
	GetIDs    []string

	UpdateErr error
	Updated   []domain.Supplier

	DeleteErr  error
	DeletedIDs []string

	ListResult  domain.Page[domain.Supplier]
	ListErr     error
	ListFilters []gateway.SupplierFilter
}

func (fake *SupplierRepository) Create(_ context.Context, supplier domain.Supplier) (domain.Supplier, error) {
	if fake.CreateErr != nil {
		return domain.Supplier{}, fake.CreateErr
	}
	supplier.ID = fmt.Sprintf("supplier-%d", len(fake.Created)+1)
	fake.Created = append(fake.Created, supplier)

	return supplier, nil
}

func (fake *SupplierRepository) Get(_ context.Context, id string) (domain.Supplier, error) {
	fake.GetIDs = append(fake.GetIDs, id)
	if fake.GetErr != nil {
		return domain.Supplier{}, fake.GetErr
	}

	return fake.GetResult, nil
}

func (fake *SupplierRepository) Update(_ context.Context, supplier domain.Supplier) (domain.Supplier, error) {
	if fake.UpdateErr != nil {
		return domain.Supplier{}, fake.UpdateErr
	}
	fake.Updated = append(fake.Updated, supplier)

	return supplier, nil
}

func (fake *SupplierRepository) Delete(_ context.Context, id string) error {
	if fake.DeleteErr != nil {
		return fake.DeleteErr
	}
	fake.DeletedIDs = append(fake.DeletedIDs, id)

	return nil
}

func (fake *SupplierRepository) List(_ context.Context, filter gateway.SupplierFilter) (domain.Page[domain.Supplier], error) {
	fake.ListFilters = append(fake.ListFilters, filter)
	if fake.ListErr != nil {
		return domain.Page[domain.Supplier]{}, fake.ListErr
	}

	return fake.ListResult, nil
}
