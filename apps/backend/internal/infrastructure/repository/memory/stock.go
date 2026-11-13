package memory

import (
	"cmp"
	"context"
	"slices"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

var _ gateway.StockRepository = (*StockRepository)(nil)

type StockRepository struct {
	store *Store
}

func NewStockRepository(store *Store) *StockRepository {
	return &StockRepository{store: store}
}

// GetLevel locks nothing on its own: the store lock that WithinTx holds is
// what keeps the level locked until the unit of work ends.
func (repository *StockRepository) GetLevel(ctx context.Context, itemID string) (domain.StockLevel, error) {
	var level domain.StockLevel
	err := repository.store.run(ctx, func(*tx) error {
		level = repository.levelOf(itemID)
		return nil
	})

	return level, err
}

func (repository *StockRepository) Record(ctx context.Context, movement domain.StockMovement) (domain.StockMovement, error) {
	var recorded domain.StockMovement
	err := repository.store.run(ctx, func(current *tx) error {
		if _, ok := repository.store.items[movement.ItemID]; !ok {
			return domain.ErrStockItemNotFound
		}
		if movement.SupplierID != nil {
			if _, ok := repository.store.suppliers[*movement.SupplierID]; !ok {
				return domain.ErrStockSupplierNotFound
			}
		}
		movement.ID = newID()
		appendRow(current, &repository.store.stockMovements, cloneMovement(movement))
		put(current, repository.store.stockLevels, movement.ItemID, domain.StockLevel{
			ItemID:    movement.ItemID,
			Quantity:  movement.BalanceAfter,
			UpdatedAt: movement.CreatedAt,
		})
		recorded = cloneMovement(movement)
		return nil
	})

	return recorded, err
}

func (repository *StockRepository) ListLevels(ctx context.Context, filter gateway.StockLevelFilter) (domain.Page[domain.ItemStock], error) {
	page := domain.Page[domain.ItemStock]{Items: []domain.ItemStock{}, Page: filter.Page.Page, Limit: filter.Page.Limit}
	err := repository.store.run(ctx, func(*tx) error {
		search := domain.NormalizeText(filter.Search)
		var matches []domain.ItemStock
		for _, item := range repository.store.items {
			if filter.CategoryID != "" && item.CategoryID != filter.CategoryID || !itemMatches(item, search) {
				continue
			}
			if len(filter.ItemIDs) > 0 && !slices.Contains(filter.ItemIDs, item.ID) {
				continue
			}
			itemStock := domain.ItemStock{Item: item, Level: repository.levelOf(item.ID)}
			if !filter.LowStock || itemStock.IsLow() {
				matches = append(matches, itemStock)
			}
		}
		slices.SortFunc(matches, func(a, b domain.ItemStock) int {
			return cmp.Or(
				cmp.Compare(domain.NormalizeText(a.Item.Name), domain.NormalizeText(b.Item.Name)),
				cmp.Compare(a.Item.CodeKey(), b.Item.CodeKey()),
			)
		})

		page.Total = len(matches)
		start := min(filter.Page.Offset(), len(matches))
		end := min(start+filter.Page.Limit, len(matches))
		for _, itemStock := range matches[start:end] {
			itemStock.Item = cloneItem(itemStock.Item)
			page.Items = append(page.Items, itemStock)
		}
		return nil
	})

	return page, err
}

func (repository *StockRepository) ListMovements(ctx context.Context, filter gateway.StockMovementFilter) (domain.Page[domain.StockMovement], error) {
	page := domain.Page[domain.StockMovement]{Items: []domain.StockMovement{}, Page: filter.Page.Page, Limit: filter.Page.Limit}
	err := repository.store.run(ctx, func(*tx) error {
		var matches []domain.StockMovement
		movements := repository.store.stockMovements
		for i := len(movements) - 1; i >= 0; i-- {
			movement := movements[i]
			if (filter.ItemID == "" || movement.ItemID == filter.ItemID) && (filter.Type == "" || movement.Type == filter.Type) {
				matches = append(matches, movement)
			}
		}
		// Stable, so movements of the same instant keep the reverse
		// recording order.
		slices.SortStableFunc(matches, func(a, b domain.StockMovement) int {
			return b.CreatedAt.Compare(a.CreatedAt)
		})

		page.Total = len(matches)
		start := min(filter.Page.Offset(), len(matches))
		end := min(start+filter.Page.Limit, len(matches))
		for _, movement := range matches[start:end] {
			page.Items = append(page.Items, cloneMovement(movement))
		}
		return nil
	})

	return page, err
}

func (repository *StockRepository) levelOf(itemID string) domain.StockLevel {
	if level, ok := repository.store.stockLevels[itemID]; ok {
		return level
	}

	return domain.StockLevel{ItemID: itemID}
}

func cloneMovement(movement domain.StockMovement) domain.StockMovement {
	movement.Reason = cloneString(movement.Reason)
	movement.Reference = cloneString(movement.Reference)
	movement.SupplierID = cloneString(movement.SupplierID)
	if movement.UnitCost != nil {
		cost := *movement.UnitCost
		movement.UnitCost = &cost
	}

	return movement
}
