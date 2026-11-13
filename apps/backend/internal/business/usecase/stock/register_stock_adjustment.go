package stock

import (
	"context"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type RegisterStockAdjustmentInput struct {
	Actor      domain.Actor
	Adjustment domain.StockAdjustment
}

// RegisterStockAdjustment corrects the stock of an item by a signed
// quantity, with a reason.
type RegisterStockAdjustment interface {
	Execute(ctx context.Context, input RegisterStockAdjustmentInput) (domain.StockMovement, error)
}

type registerStockAdjustment struct {
	tx    gateway.TxManager
	stock gateway.StockRepository
	items gateway.ItemRepository
}

func NewRegisterStockAdjustment(tx gateway.TxManager, stock gateway.StockRepository, items gateway.ItemRepository) RegisterStockAdjustment {
	return &registerStockAdjustment{tx: tx, stock: stock, items: items}
}

func (useCase *registerStockAdjustment) Execute(ctx context.Context, input RegisterStockAdjustmentInput) (domain.StockMovement, error) {
	if err := input.Actor.Authorize(domain.PermissionManageStock); err != nil {
		return domain.StockMovement{}, err
	}

	adjustment, err := domain.NewStockAdjustment(input.Adjustment)
	if err != nil {
		return domain.StockMovement{}, err
	}

	now := time.Now().UTC()
	var recorded domain.StockMovement
	err = useCase.tx.WithinTx(ctx, func(ctx context.Context) error {
		item, err := useCase.items.Get(ctx, adjustment.ItemID)
		if err != nil {
			return referenced(err, domain.ErrItemNotFound, domain.ErrStockItemNotFound)
		}
		level, err := useCase.stock.GetLevel(ctx, item.ID)
		if err != nil {
			return err
		}
		movement, err := level.Adjust(item, adjustment, input.Actor, now)
		if err != nil {
			return err
		}
		recorded, err = useCase.stock.Record(ctx, movement)
		return err
	})
	if err != nil {
		return domain.StockMovement{}, domain.FromStorage(err)
	}

	return recorded, nil
}
