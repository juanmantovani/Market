package stock

import (
	"context"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type RegisterInitialStockInput struct {
	Actor    domain.Actor
	ItemID   string
	Quantity domain.Quantity
}

// RegisterInitialStock opens the stock of an item that has no movements yet.
type RegisterInitialStock interface {
	Execute(ctx context.Context, input RegisterInitialStockInput) (domain.StockMovement, error)
}

type registerInitialStock struct {
	tx    gateway.TxManager
	stock gateway.StockRepository
	items gateway.ItemRepository
}

func NewRegisterInitialStock(tx gateway.TxManager, stock gateway.StockRepository, items gateway.ItemRepository) RegisterInitialStock {
	return &registerInitialStock{tx: tx, stock: stock, items: items}
}

func (useCase *registerInitialStock) Execute(ctx context.Context, input RegisterInitialStockInput) (domain.StockMovement, error) {
	if err := input.Actor.Authorize(domain.PermissionManageStock); err != nil {
		return domain.StockMovement{}, err
	}

	now := time.Now().UTC()
	var recorded domain.StockMovement
	err := useCase.tx.WithinTx(ctx, func(ctx context.Context) error {
		item, err := useCase.items.Get(ctx, input.ItemID)
		if err != nil {
			return referenced(err, domain.ErrItemNotFound, domain.ErrStockItemNotFound)
		}
		level, err := useCase.stock.GetLevel(ctx, item.ID)
		if err != nil {
			return err
		}
		movement, err := level.Initial(item, input.Quantity, input.Actor, now)
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
