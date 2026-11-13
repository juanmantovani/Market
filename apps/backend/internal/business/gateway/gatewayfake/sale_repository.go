package gatewayfake

import (
	"context"
	"fmt"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

// SaleRepository is a fake gateway.SaleRepository for use case tests. It
// records every call and returns the configured results. Create numbers the
// sales it saves from 1, and GetByIdempotencyKey finds them, so a retried
// request sees the sale the first one created.
type SaleRepository struct {
	CreateErr error
	Created   []domain.Sale

	GetResult domain.Sale
	GetErr    error
	GetIDs    []string

	GetByKeyErr error
	Keys        []string

	VoidErr error
	Voided  []domain.Sale

	// ListPages, when set, are returned one per call instead of ListResult.
	ListPages   []domain.Page[domain.Sale]
	ListResult  domain.Page[domain.Sale]
	ListErr     error
	ListFilters []gateway.SaleFilter
}

func (fake *SaleRepository) Create(_ context.Context, sale domain.Sale) (domain.Sale, error) {
	if fake.CreateErr != nil {
		return domain.Sale{}, fake.CreateErr
	}
	sale.ID = fmt.Sprintf("sale-%d", len(fake.Created)+1)
	sale.Number = int64(len(fake.Created) + 1)
	fake.Created = append(fake.Created, sale)

	return sale, nil
}

func (fake *SaleRepository) Get(_ context.Context, id string) (domain.Sale, error) {
	fake.GetIDs = append(fake.GetIDs, id)
	if fake.GetErr != nil {
		return domain.Sale{}, fake.GetErr
	}

	return fake.GetResult, nil
}

func (fake *SaleRepository) GetByIdempotencyKey(_ context.Context, actorID, key string) (domain.Sale, error) {
	fake.Keys = append(fake.Keys, key)
	if fake.GetByKeyErr != nil {
		return domain.Sale{}, fake.GetByKeyErr
	}
	for _, sale := range fake.Created {
		if sale.ActorID == actorID && sale.IdempotencyKey == key {
			return sale, nil
		}
	}

	return domain.Sale{}, domain.ErrSaleNotFound
}

func (fake *SaleRepository) Void(_ context.Context, sale domain.Sale) (domain.Sale, error) {
	if fake.VoidErr != nil {
		return domain.Sale{}, fake.VoidErr
	}
	fake.Voided = append(fake.Voided, sale)

	return sale, nil
}

func (fake *SaleRepository) List(_ context.Context, filter gateway.SaleFilter) (domain.Page[domain.Sale], error) {
	fake.ListFilters = append(fake.ListFilters, filter)
	if fake.ListErr != nil {
		return domain.Page[domain.Sale]{}, fake.ListErr
	}

	if call := len(fake.ListFilters) - 1; call < len(fake.ListPages) {
		return fake.ListPages[call], nil
	}

	return fake.ListResult, nil
}
