package stock

import (
	"context"
	"errors"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type RegisterStockEntryInput struct {
	Actor domain.Actor
	Entry domain.StockEntry
}

// RegisterStockEntry adds received merchandise to the stock, one movement
// per line, and updates the cost of every item the entry gives a cost for.
// Either every line is recorded or none is.
type RegisterStockEntry interface {
	Execute(ctx context.Context, input RegisterStockEntryInput) ([]domain.StockMovement, error)
}

type registerStockEntry struct {
	tx        gateway.TxManager
	stock     gateway.StockRepository
	items     gateway.ItemRepository
	suppliers gateway.SupplierRepository
}

func NewRegisterStockEntry(tx gateway.TxManager, stock gateway.StockRepository, items gateway.ItemRepository, suppliers gateway.SupplierRepository) RegisterStockEntry {
	return &registerStockEntry{tx: tx, stock: stock, items: items, suppliers: suppliers}
}

func (useCase *registerStockEntry) Execute(ctx context.Context, input RegisterStockEntryInput) ([]domain.StockMovement, error) {
	if err := input.Actor.Authorize(domain.PermissionManageStock); err != nil {
		return nil, err
	}

	entry, err := domain.NewStockEntry(input.Entry)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	var movements []domain.StockMovement
	err = useCase.tx.WithinTx(ctx, func(ctx context.Context) error {
		movements = make([]domain.StockMovement, 0, len(entry.Lines))
		if entry.SupplierID != nil {
			if _, err := useCase.suppliers.Get(ctx, *entry.SupplierID); err != nil {
				return referenced(err, domain.ErrSupplierNotFound, domain.ErrStockSupplierNotFound)
			}
		}
		for _, line := range entry.Lines {
			movement, err := useCase.receive(ctx, entry, line, input.Actor, now)
			if err != nil {
				return err
			}
			movements = append(movements, movement)
		}
		return nil
	})
	if err != nil {
		return nil, domain.FromStorage(err)
	}

	return movements, nil
}

func (useCase *registerStockEntry) receive(ctx context.Context, entry domain.StockEntry, line domain.StockEntryLine, actor domain.Actor, now time.Time) (domain.StockMovement, error) {
	item, err := useCase.items.Get(ctx, line.ItemID)
	if err != nil {
		return domain.StockMovement{}, referenced(err, domain.ErrItemNotFound, domain.ErrStockItemNotFound)
	}
	level, err := useCase.stock.GetLevel(ctx, item.ID)
	if err != nil {
		return domain.StockMovement{}, err
	}
	movement, err := level.Receive(item, entry, line, actor, now)
	if err != nil {
		return domain.StockMovement{}, err
	}
	recorded, err := useCase.stock.Record(ctx, movement)
	if err != nil {
		return domain.StockMovement{}, err
	}

	costed, changed, err := line.UpdateCost(item, now)
	if err != nil {
		return domain.StockMovement{}, err
	}
	if changed {
		if _, err := useCase.items.Update(ctx, costed); err != nil {
			return domain.StockMovement{}, err
		}
	}

	return recorded, nil
}

// referenced replaces the not-found error of a record the request refers to
// with the error that blames the request instead of the URL.
func referenced(err error, notFound, replacement *domain.Error) error {
	if errors.Is(err, notFound) {
		return replacement
	}

	return err
}
