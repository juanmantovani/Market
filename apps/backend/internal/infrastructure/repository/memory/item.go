package memory

import (
	"cmp"
	"context"
	"slices"
	"strings"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

var _ gateway.ItemRepository = (*ItemRepository)(nil)

type ItemRepository struct {
	store *Store
}

func NewItemRepository(store *Store) *ItemRepository {
	return &ItemRepository{store: store}
}

func (repository *ItemRepository) Create(ctx context.Context, item domain.Item) (domain.Item, error) {
	var created domain.Item
	err := repository.store.run(ctx, func(current *tx) error {
		item.ID = newID()
		if err := repository.checkConstraints(item); err != nil {
			return err
		}
		put(current, repository.store.items, item.ID, cloneItem(item))
		created = cloneItem(item)
		return nil
	})

	return created, err
}

func (repository *ItemRepository) Get(ctx context.Context, id string) (domain.Item, error) {
	var found domain.Item
	err := repository.store.run(ctx, func(*tx) error {
		item, ok := repository.store.items[id]
		if !ok {
			return domain.ErrItemNotFound
		}
		found = cloneItem(item)
		return nil
	})

	return found, err
}

func (repository *ItemRepository) Update(ctx context.Context, item domain.Item) (domain.Item, error) {
	var updated domain.Item
	err := repository.store.run(ctx, func(current *tx) error {
		if _, ok := repository.store.items[item.ID]; !ok {
			return domain.ErrItemNotFound
		}
		if err := repository.checkConstraints(item); err != nil {
			return err
		}
		put(current, repository.store.items, item.ID, cloneItem(item))
		updated = cloneItem(item)
		return nil
	})

	return updated, err
}

func (repository *ItemRepository) Delete(ctx context.Context, id string) error {
	return repository.store.run(ctx, func(current *tx) error {
		if _, ok := repository.store.items[id]; !ok {
			return domain.ErrItemNotFound
		}
		if _, ok := repository.store.stockLevels[id]; ok {
			return domain.ErrItemInUse
		}
		remove(current, repository.store.items, id)
		return nil
	})
}

func (repository *ItemRepository) List(ctx context.Context, filter gateway.ItemFilter) (domain.Page[domain.Item], error) {
	page := domain.Page[domain.Item]{Items: []domain.Item{}, Page: filter.Page.Page, Limit: filter.Page.Limit}
	err := repository.store.run(ctx, func(*tx) error {
		search := domain.NormalizeText(filter.Search)
		var matches []domain.Item
		for _, item := range repository.store.items {
			if (filter.CategoryID == "" || item.CategoryID == filter.CategoryID) && itemMatches(item, search) && (!filter.LowStock || domain.IsLowStock(repository.store.stockLevels[item.ID].Quantity, item.MinStock)) {
				matches = append(matches, item)
			}
		}
		slices.SortFunc(matches, func(a, b domain.Item) int {
			return cmp.Or(
				cmp.Compare(domain.NormalizeText(a.Name), domain.NormalizeText(b.Name)),
				cmp.Compare(a.CodeKey(), b.CodeKey()),
			)
		})

		page.Total = len(matches)
		start := min(filter.Page.Offset(), len(matches))
		end := min(start+filter.Page.Limit, len(matches))
		for _, item := range matches[start:end] {
			page.Items = append(page.Items, cloneItem(item))
		}
		return nil
	})

	return page, err
}

// checkConstraints enforces what unique indexes on the code and barcode and
// foreign keys to categories and suppliers enforce in PostgreSQL.
func (repository *ItemRepository) checkConstraints(item domain.Item) error {
	if _, ok := repository.store.categories[item.CategoryID]; !ok {
		return domain.ErrItemCategoryNotFound
	}
	for _, supplierID := range item.SupplierIDs {
		if _, ok := repository.store.suppliers[supplierID]; !ok {
			return domain.ErrItemSupplierNotFound
		}
	}
	codeKey := item.CodeKey()
	for id, other := range repository.store.items {
		if id == item.ID {
			continue
		}
		if other.CodeKey() == codeKey {
			return domain.ErrItemCodeInUse
		}
		if item.Barcode != nil && other.Barcode != nil && *item.Barcode == *other.Barcode {
			return domain.ErrItemBarcodeInUse
		}
	}

	return nil
}

func itemMatches(item domain.Item, search string) bool {
	if search == "" {
		return true
	}

	return strings.Contains(domain.NormalizeText(item.Name), search) ||
		strings.Contains(item.CodeKey(), search) ||
		(item.Barcode != nil && strings.Contains(domain.NormalizeText(*item.Barcode), search))
}

func cloneItem(item domain.Item) domain.Item {
	item.Barcode = cloneString(item.Barcode)
	item.Description = cloneString(item.Description)
	item.SupplierIDs = slices.Clone(item.SupplierIDs)
	if item.Cost != nil {
		cost := *item.Cost
		item.Cost = &cost
	}

	return item
}

func cloneString(value *string) *string {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
