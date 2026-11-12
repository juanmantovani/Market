package stock

import (
	"context"
	"strings"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type ListStockMovementsInput struct {
	Actor  domain.Actor
	ItemID string
	Type   domain.MovementType
	Page   int
	Limit  int
}

// ListStockMovements lists stock movements newest first, optionally only
// those of one item or one type.
type ListStockMovements interface {
	Execute(ctx context.Context, input ListStockMovementsInput) (domain.Page[domain.StockMovement], error)
}

type listStockMovements struct {
	repository gateway.StockRepository
}

func NewListStockMovements(repository gateway.StockRepository) ListStockMovements {
	return &listStockMovements{repository: repository}
}

func (useCase *listStockMovements) Execute(ctx context.Context, input ListStockMovementsInput) (domain.Page[domain.StockMovement], error) {
	if err := input.Actor.Authorize(domain.PermissionViewStock); err != nil {
		return domain.Page[domain.StockMovement]{}, err
	}

	if input.Type != "" && !input.Type.IsValid() {
		return domain.Page[domain.StockMovement]{}, domain.ErrStockMovementTypeInvalid
	}
	page, err := domain.NewPageRequest(input.Page, input.Limit)
	if err != nil {
		return domain.Page[domain.StockMovement]{}, err
	}

	result, err := useCase.repository.ListMovements(ctx, gateway.StockMovementFilter{
		ItemID: strings.TrimSpace(input.ItemID),
		Type:   input.Type,
		Page:   page,
	})
	if err != nil {
		return domain.Page[domain.StockMovement]{}, domain.FromStorage(err)
	}

	return result, nil
}
