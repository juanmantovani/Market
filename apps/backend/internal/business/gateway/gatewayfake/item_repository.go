package gatewayfake

import (
	"context"
	"fmt"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

// ItemRepository is a fake gateway.ItemRepository for use case tests. It
// records every call and returns the configured results.
type ItemRepository struct {
	CreateErr error
	Created   []domain.Item

	GetResult domain.Item
	GetErr    error
	GetIDs    []string
	// Stored, when set, answers Get by ID instead of GetResult, with
	// ErrItemNotFound for any other ID.
	Stored map[string]domain.Item

	UpdateErr error
	Updated   []domain.Item

	DeleteErr  error
	DeletedIDs []string

	ListResult  domain.Page[domain.Item]
	ListErr     error
	ListFilters []gateway.ItemFilter
}

func (fake *ItemRepository) Create(_ context.Context, item domain.Item) (domain.Item, error) {
	if fake.CreateErr != nil {
		return domain.Item{}, fake.CreateErr
	}
	item.ID = fmt.Sprintf("item-%d", len(fake.Created)+1)
	fake.Created = append(fake.Created, item)

	return item, nil
}

func (fake *ItemRepository) Get(_ context.Context, id string) (domain.Item, error) {
	fake.GetIDs = append(fake.GetIDs, id)
	if fake.GetErr != nil {
		return domain.Item{}, fake.GetErr
	}
	if fake.Stored != nil {
		item, ok := fake.Stored[id]
		if !ok {
			return domain.Item{}, domain.ErrItemNotFound
		}
		return item, nil
	}

	return fake.GetResult, nil
}

func (fake *ItemRepository) Update(_ context.Context, item domain.Item) (domain.Item, error) {
	if fake.UpdateErr != nil {
		return domain.Item{}, fake.UpdateErr
	}
	fake.Updated = append(fake.Updated, item)

	return item, nil
}

func (fake *ItemRepository) Delete(_ context.Context, id string) error {
	if fake.DeleteErr != nil {
		return fake.DeleteErr
	}
	fake.DeletedIDs = append(fake.DeletedIDs, id)

	return nil
}

func (fake *ItemRepository) List(_ context.Context, filter gateway.ItemFilter) (domain.Page[domain.Item], error) {
	fake.ListFilters = append(fake.ListFilters, filter)
	if fake.ListErr != nil {
		return domain.Page[domain.Item]{}, fake.ListErr
	}

	return fake.ListResult, nil
}
