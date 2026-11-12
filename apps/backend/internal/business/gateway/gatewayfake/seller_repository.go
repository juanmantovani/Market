package gatewayfake

import (
	"context"
	"fmt"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

// SellerRepository is a fake gateway.SellerRepository for use case
// tests. It records every call and returns the configured results.
type SellerRepository struct {
	CreateErr error
	Created   []domain.Seller

	GetResult domain.Seller
	GetErr    error
	GetIDs    []string

	UpdateErr error
	Updated   []domain.Seller

	DeleteErr  error
	DeletedIDs []string

	ListResult  domain.Page[domain.Seller]
	ListErr     error
	ListFilters []gateway.SellerFilter
}

func (fake *SellerRepository) Create(_ context.Context, seller domain.Seller) (domain.Seller, error) {
	if fake.CreateErr != nil {
		return domain.Seller{}, fake.CreateErr
	}
	seller.ID = fmt.Sprintf("seller-%d", len(fake.Created)+1)
	fake.Created = append(fake.Created, seller)

	return seller, nil
}

func (fake *SellerRepository) Get(_ context.Context, id string) (domain.Seller, error) {
	fake.GetIDs = append(fake.GetIDs, id)
	if fake.GetErr != nil {
		return domain.Seller{}, fake.GetErr
	}

	return fake.GetResult, nil
}

func (fake *SellerRepository) Update(_ context.Context, seller domain.Seller) (domain.Seller, error) {
	if fake.UpdateErr != nil {
		return domain.Seller{}, fake.UpdateErr
	}
	fake.Updated = append(fake.Updated, seller)

	return seller, nil
}

func (fake *SellerRepository) Delete(_ context.Context, id string) error {
	if fake.DeleteErr != nil {
		return fake.DeleteErr
	}
	fake.DeletedIDs = append(fake.DeletedIDs, id)

	return nil
}

func (fake *SellerRepository) List(_ context.Context, filter gateway.SellerFilter) (domain.Page[domain.Seller], error) {
	fake.ListFilters = append(fake.ListFilters, filter)
	if fake.ListErr != nil {
		return domain.Page[domain.Seller]{}, fake.ListErr
	}

	return fake.ListResult, nil
}
