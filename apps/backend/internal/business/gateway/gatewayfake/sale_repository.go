package gatewayfake

import (
	"context"
	"fmt"

	"market/backend/internal/business/domain"
)

// SaleRepository is a fake gateway.SaleRepository for use case tests. It
// records every call and returns the configured results. Create numbers the
// sales it saves from 1, and GetByIdempotencyKey finds them, so a retried
// request sees the sale the first one created.
type SaleRepository struct {
	CreateErr error
	Created   []domain.Sale

	GetByKeyErr error
	Keys        []string
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

func (fake *SaleRepository) GetByIdempotencyKey(_ context.Context, key string) (domain.Sale, error) {
	fake.Keys = append(fake.Keys, key)
	if fake.GetByKeyErr != nil {
		return domain.Sale{}, fake.GetByKeyErr
	}
	for _, sale := range fake.Created {
		if sale.IdempotencyKey == key {
			return sale, nil
		}
	}

	return domain.Sale{}, domain.ErrSaleNotFound
}
