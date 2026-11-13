package stock

import (
	"context"
	"strings"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type ListStockLevelsInput struct {
	Actor      domain.Actor
	Search     string
	CategoryID string
	ItemIDs    []string
	LowStock   bool
	Page       int
	Limit      int
}

// ListStockLevels lists every item with its stock, optionally only some items
// or only those whose stock reached their minimum.
type ListStockLevels interface {
	Execute(ctx context.Context, input ListStockLevelsInput) (domain.Page[domain.ItemStock], error)
}

type listStockLevels struct {
	repository gateway.StockRepository
}

func NewListStockLevels(repository gateway.StockRepository) ListStockLevels {
	return &listStockLevels{repository: repository}
}

func (useCase *listStockLevels) Execute(ctx context.Context, input ListStockLevelsInput) (domain.Page[domain.ItemStock], error) {
	if err := input.Actor.Authorize(domain.PermissionViewStock); err != nil {
		return domain.Page[domain.ItemStock]{}, err
	}

	page, err := domain.NewPageRequest(input.Page, input.Limit)
	if err != nil {
		return domain.Page[domain.ItemStock]{}, err
	}

	result, err := useCase.repository.ListLevels(ctx, gateway.StockLevelFilter{
		Search:     strings.TrimSpace(input.Search),
		CategoryID: strings.TrimSpace(input.CategoryID),
		ItemIDs:    cleanIDs(input.ItemIDs),
		LowStock:   input.LowStock,
		Page:       page,
	})
	if err != nil {
		return domain.Page[domain.ItemStock]{}, domain.FromStorage(err)
	}

	return result, nil
}

// cleanIDs trims the IDs and drops the empty ones; it returns nil when none
// is left, which means no filter.
func cleanIDs(ids []string) []string {
	var cleaned []string
	for _, id := range ids {
		if id = strings.TrimSpace(id); id != "" {
			cleaned = append(cleaned, id)
		}
	}

	return cleaned
}
