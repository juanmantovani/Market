package memory

import (
	"context"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

var _ gateway.SaleRepository = (*SaleRepository)(nil)

type SaleRepository struct {
	store *Store
}

func NewSaleRepository(store *Store) *SaleRepository {
	return &SaleRepository{store: store}
}

func (repository *SaleRepository) Create(ctx context.Context, sale domain.Sale) (domain.Sale, error) {
	var created domain.Sale
	err := repository.store.run(ctx, func(current *tx) error {
		if err := repository.checkReferences(sale); err != nil {
			return err
		}
		sale.ID = newID()
		set(current, &repository.store.lastSaleNumber, repository.store.lastSaleNumber+1)
		sale.Number = repository.store.lastSaleNumber
		put(current, repository.store.sales, sale.ID, cloneSale(sale))
		created = cloneSale(sale)
		return nil
	})

	return created, err
}

func (repository *SaleRepository) GetByIdempotencyKey(ctx context.Context, key string) (domain.Sale, error) {
	var found domain.Sale
	err := repository.store.run(ctx, func(*tx) error {
		for _, sale := range repository.store.sales {
			if sale.IdempotencyKey == key {
				found = cloneSale(sale)
				return nil
			}
		}
		return domain.ErrSaleNotFound
	})

	return found, err
}

// checkReferences enforces what a unique index on the idempotency key and
// the foreign keys to customers, sellers and items enforce in PostgreSQL.
func (repository *SaleRepository) checkReferences(sale domain.Sale) error {
	for _, other := range repository.store.sales {
		if other.IdempotencyKey == sale.IdempotencyKey {
			return domain.ErrIdempotencyKeyConflict
		}
	}
	if _, ok := repository.store.customers[sale.Customer.ID]; !ok {
		return domain.ErrSaleCustomerNotFound
	}
	if sale.Seller != nil {
		if _, ok := repository.store.sellers[sale.Seller.ID]; !ok {
			return domain.ErrSaleSellerNotFound
		}
	}
	for _, line := range sale.Lines {
		if _, ok := repository.store.items[line.ItemID]; !ok {
			return domain.ErrSaleItemNotFound
		}
	}

	return nil
}

// hasSaleWhere reports whether any stored sale matches. The caller holds the
// store lock.
func (store *Store) hasSaleWhere(matches func(domain.Sale) bool) bool {
	for _, sale := range store.sales {
		if matches(sale) {
			return true
		}
	}

	return false
}

func cloneSale(sale domain.Sale) domain.Sale {
	if sale.Seller != nil {
		seller := *sale.Seller
		sale.Seller = &seller
	}
	sale.Lines = append([]domain.SaleLine(nil), sale.Lines...)
	payments := make([]domain.Payment, len(sale.Payments))
	for i, payment := range sale.Payments {
		payment.Reference = cloneString(payment.Reference)
		payments[i] = payment
	}
	sale.Payments = payments

	return sale
}
