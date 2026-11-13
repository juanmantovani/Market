package memory

import (
	"cmp"
	"context"
	"slices"
	"strings"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

var _ gateway.SellerRepository = (*SellerRepository)(nil)

type SellerRepository struct {
	store *Store
}

func NewSellerRepository(store *Store) *SellerRepository {
	return &SellerRepository{store: store}
}

func (repository *SellerRepository) Create(ctx context.Context, seller domain.Seller) (domain.Seller, error) {
	var created domain.Seller
	err := repository.store.run(ctx, func(current *tx) error {
		seller.ID = newID()
		if err := repository.checkCodeIsFree(seller); err != nil {
			return err
		}
		put(current, repository.store.sellers, seller.ID, cloneSeller(seller))
		created = cloneSeller(seller)
		return nil
	})

	return created, err
}

func (repository *SellerRepository) Get(ctx context.Context, id string) (domain.Seller, error) {
	var found domain.Seller
	err := repository.store.run(ctx, func(*tx) error {
		seller, ok := repository.store.sellers[id]
		if !ok {
			return domain.ErrSellerNotFound
		}
		found = cloneSeller(seller)
		return nil
	})

	return found, err
}

func (repository *SellerRepository) Update(ctx context.Context, seller domain.Seller) (domain.Seller, error) {
	var updated domain.Seller
	err := repository.store.run(ctx, func(current *tx) error {
		if _, ok := repository.store.sellers[seller.ID]; !ok {
			return domain.ErrSellerNotFound
		}
		if err := repository.checkCodeIsFree(seller); err != nil {
			return err
		}
		put(current, repository.store.sellers, seller.ID, cloneSeller(seller))
		updated = cloneSeller(seller)
		return nil
	})

	return updated, err
}

func (repository *SellerRepository) Delete(ctx context.Context, id string) error {
	return repository.store.run(ctx, func(current *tx) error {
		if _, ok := repository.store.sellers[id]; !ok {
			return domain.ErrSellerNotFound
		}
		if repository.store.hasSaleWhere(func(sale domain.Sale) bool { return sale.Seller != nil && sale.Seller.ID == id }) {
			return domain.ErrSellerInUse
		}
		remove(current, repository.store.sellers, id)
		return nil
	})
}

func (repository *SellerRepository) List(ctx context.Context, filter gateway.SellerFilter) (domain.Page[domain.Seller], error) {
	page := domain.Page[domain.Seller]{Items: []domain.Seller{}, Page: filter.Page.Page, Limit: filter.Page.Limit}
	err := repository.store.run(ctx, func(*tx) error {
		search := domain.NormalizeText(filter.Search)
		var matches []domain.Seller
		for _, seller := range repository.store.sellers {
			if search == "" || strings.Contains(domain.NormalizeText(seller.Name), search) ||
				strings.Contains(strings.ToLower(seller.Code), search) {
				matches = append(matches, seller)
			}
		}
		slices.SortFunc(matches, func(a, b domain.Seller) int {
			return cmp.Or(cmp.Compare(domain.NormalizeText(a.Name), domain.NormalizeText(b.Name)), cmp.Compare(a.ID, b.ID))
		})

		page.Total = len(matches)
		start := min(filter.Page.Offset(), len(matches))
		end := min(start+filter.Page.Limit, len(matches))
		for _, seller := range matches[start:end] {
			page.Items = append(page.Items, cloneSeller(seller))
		}
		return nil
	})

	return page, err
}

// checkCodeIsFree enforces what a unique index on the code enforces in
// PostgreSQL. Codes are already upper case, so equal means the same code.
func (repository *SellerRepository) checkCodeIsFree(seller domain.Seller) error {
	for id, other := range repository.store.sellers {
		if id != seller.ID && other.Code == seller.Code {
			return domain.ErrSellerCodeInUse
		}
	}

	return nil
}

func cloneSeller(seller domain.Seller) domain.Seller {
	if seller.CommissionRate != nil {
		rate := *seller.CommissionRate
		seller.CommissionRate = &rate
	}

	return seller
}
