package memory

import (
	"cmp"
	"context"
	"slices"
	"strings"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

var _ gateway.SupplierRepository = (*SupplierRepository)(nil)

type SupplierRepository struct {
	store *Store
}

func NewSupplierRepository(store *Store) *SupplierRepository {
	return &SupplierRepository{store: store}
}

func (repository *SupplierRepository) Create(ctx context.Context, supplier domain.Supplier) (domain.Supplier, error) {
	var created domain.Supplier
	err := repository.store.run(ctx, func(current *tx) error {
		supplier.ID = newID()
		if err := repository.checkCUITIsFree(supplier); err != nil {
			return err
		}
		put(current, repository.store.suppliers, supplier.ID, cloneSupplier(supplier))
		created = cloneSupplier(supplier)
		return nil
	})

	return created, err
}

func (repository *SupplierRepository) Get(ctx context.Context, id string) (domain.Supplier, error) {
	var found domain.Supplier
	err := repository.store.run(ctx, func(*tx) error {
		supplier, ok := repository.store.suppliers[id]
		if !ok {
			return domain.ErrSupplierNotFound
		}
		found = cloneSupplier(supplier)
		return nil
	})

	return found, err
}

func (repository *SupplierRepository) Update(ctx context.Context, supplier domain.Supplier) (domain.Supplier, error) {
	var updated domain.Supplier
	err := repository.store.run(ctx, func(current *tx) error {
		if _, ok := repository.store.suppliers[supplier.ID]; !ok {
			return domain.ErrSupplierNotFound
		}
		if err := repository.checkCUITIsFree(supplier); err != nil {
			return err
		}
		put(current, repository.store.suppliers, supplier.ID, cloneSupplier(supplier))
		updated = cloneSupplier(supplier)
		return nil
	})

	return updated, err
}

func (repository *SupplierRepository) Delete(ctx context.Context, id string) error {
	return repository.store.run(ctx, func(current *tx) error {
		if _, ok := repository.store.suppliers[id]; !ok {
			return domain.ErrSupplierNotFound
		}
		for _, item := range repository.store.items {
			if slices.Contains(item.SupplierIDs, id) {
				return domain.ErrSupplierInUse
			}
		}
		for _, movement := range repository.store.stockMovements {
			if movement.SupplierID != nil && *movement.SupplierID == id {
				return domain.ErrSupplierHasStockEntries
			}
		}
		remove(current, repository.store.suppliers, id)
		return nil
	})
}

func (repository *SupplierRepository) List(ctx context.Context, filter gateway.SupplierFilter) (domain.Page[domain.Supplier], error) {
	page := domain.Page[domain.Supplier]{Items: []domain.Supplier{}, Page: filter.Page.Page, Limit: filter.Page.Limit}
	err := repository.store.run(ctx, func(*tx) error {
		search := domain.NormalizeText(filter.Search)
		searchDigits := strings.Map(keepDigit, filter.Search)
		var matches []domain.Supplier
		for _, supplier := range repository.store.suppliers {
			if supplierMatches(supplier, search, searchDigits) {
				matches = append(matches, supplier)
			}
		}
		slices.SortFunc(matches, func(a, b domain.Supplier) int {
			return cmp.Or(
				cmp.Compare(domain.NormalizeText(a.BusinessName), domain.NormalizeText(b.BusinessName)),
				cmp.Compare(a.ID, b.ID),
			)
		})

		page.Total = len(matches)
		start := min(filter.Page.Offset(), len(matches))
		end := min(start+filter.Page.Limit, len(matches))
		for _, supplier := range matches[start:end] {
			page.Items = append(page.Items, cloneSupplier(supplier))
		}
		return nil
	})

	return page, err
}

// checkCUITIsFree enforces what a partial unique index on the CUIT (when
// there is one) enforces in PostgreSQL.
func (repository *SupplierRepository) checkCUITIsFree(supplier domain.Supplier) error {
	if supplier.CUIT == nil {
		return nil
	}
	for id, other := range repository.store.suppliers {
		if id != supplier.ID && other.CUIT != nil && *other.CUIT == *supplier.CUIT {
			return domain.ErrSupplierCUITInUse
		}
	}

	return nil
}

func supplierMatches(supplier domain.Supplier, search, searchDigits string) bool {
	if search == "" {
		return true
	}
	if searchDigits != "" && supplier.CUIT != nil && strings.Contains(*supplier.CUIT, searchDigits) {
		return true
	}
	if strings.Contains(domain.NormalizeText(supplier.BusinessName), search) {
		return true
	}
	for _, text := range []*string{supplier.ContactName, supplier.Email, supplier.Phone} {
		if text != nil && strings.Contains(domain.NormalizeText(*text), search) {
			return true
		}
	}

	return false
}

func cloneSupplier(supplier domain.Supplier) domain.Supplier {
	supplier.CUIT = cloneString(supplier.CUIT)
	supplier.ContactName = cloneString(supplier.ContactName)
	supplier.Phone = cloneString(supplier.Phone)
	supplier.Email = cloneString(supplier.Email)
	supplier.Address = cloneString(supplier.Address)
	supplier.Notes = cloneString(supplier.Notes)

	return supplier
}
