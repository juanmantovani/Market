package sales

import (
	"context"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type VoidSaleInput struct {
	Actor  domain.Actor
	ID     string
	Reason string
}

// VoidSale voids a completed sale and puts what it sold back in the stock
// with compensating movements: the void and its movements are saved
// together or not at all.
type VoidSale interface {
	Execute(ctx context.Context, input VoidSaleInput) (domain.Sale, error)
}

type voidSale struct {
	tx    gateway.TxManager
	sales gateway.SaleRepository
	stock gateway.StockRepository
}

func NewVoidSale(tx gateway.TxManager, sales gateway.SaleRepository, stock gateway.StockRepository) VoidSale {
	return &voidSale{tx: tx, sales: sales, stock: stock}
}

func (useCase *voidSale) Execute(ctx context.Context, input VoidSaleInput) (domain.Sale, error) {
	if err := input.Actor.Authorize(domain.PermissionVoidSales); err != nil {
		return domain.Sale{}, err
	}

	now := time.Now().UTC()
	var voided domain.Sale
	err := useCase.tx.WithinTx(ctx, func(ctx context.Context) error {
		sale, err := useCase.sales.Get(ctx, input.ID)
		if err != nil {
			return err
		}
		changed, err := sale.Void(input.Reason, input.Actor, now)
		if err != nil {
			return err
		}
		if voided, err = useCase.sales.Void(ctx, changed); err != nil {
			return err
		}
		for _, line := range voided.Lines {
			if err := useCase.putBack(ctx, voided, line); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return domain.Sale{}, domain.FromStorage(err)
	}

	return voided, nil
}

func (useCase *voidSale) putBack(ctx context.Context, voided domain.Sale, line domain.SaleLine) error {
	level, err := useCase.stock.GetLevel(ctx, line.ItemID)
	if err != nil {
		return err
	}
	movement, err := level.VoidSale(voided, line)
	if err != nil {
		return err
	}
	_, err = useCase.stock.Record(ctx, movement)

	return err
}
