package gatewayfake

import (
	"context"
	"fmt"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

// CategoryRepository is a fake gateway.CategoryRepository for use case
// tests. It records every call and returns the configured results; it
// doesn't enforce the contract (that's what repositorytest checks against
// the real adapters).
type CategoryRepository struct {
	CreateErr error
	Created   []domain.Category

	GetResult domain.Category
	GetErr    error
	GetIDs    []string

	UpdateErr error
	Updated   []domain.Category

	ListResult  domain.Page[domain.Category]
	ListErr     error
	ListFilters []gateway.CategoryFilter
}

func (fake *CategoryRepository) Create(_ context.Context, category domain.Category) (domain.Category, error) {
	if fake.CreateErr != nil {
		return domain.Category{}, fake.CreateErr
	}
	category.ID = fmt.Sprintf("category-%d", len(fake.Created)+1)
	fake.Created = append(fake.Created, category)

	return category, nil
}

func (fake *CategoryRepository) Get(_ context.Context, id string) (domain.Category, error) {
	fake.GetIDs = append(fake.GetIDs, id)
	if fake.GetErr != nil {
		return domain.Category{}, fake.GetErr
	}

	return fake.GetResult, nil
}

func (fake *CategoryRepository) Update(_ context.Context, category domain.Category) (domain.Category, error) {
	if fake.UpdateErr != nil {
		return domain.Category{}, fake.UpdateErr
	}
	fake.Updated = append(fake.Updated, category)

	return category, nil
}

func (fake *CategoryRepository) List(_ context.Context, filter gateway.CategoryFilter) (domain.Page[domain.Category], error) {
	fake.ListFilters = append(fake.ListFilters, filter)
	if fake.ListErr != nil {
		return domain.Page[domain.Category]{}, fake.ListErr
	}

	return fake.ListResult, nil
}
