package sales

import (
	"context"
	"errors"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

type CreateSaleInput struct {
	Actor   domain.Actor
	Request domain.SaleRequest
}

// CreateSale confirms a sale and takes what it sells out of the stock: the
// sale and its movements are saved together or not at all. A request that
// repeats the idempotency key of an earlier one gets that same sale back.
type CreateSale interface {
	Execute(ctx context.Context, input CreateSaleInput) (domain.Sale, error)
}

type createSale struct {
	tx        gateway.TxManager
	sales     gateway.SaleRepository
	items     gateway.ItemRepository
	customers gateway.CustomerRepository
	sellers   gateway.SellerRepository
	stock     gateway.StockRepository
	policy    domain.StockPolicy
}

func NewCreateSale(
	tx gateway.TxManager,
	sales gateway.SaleRepository,
	items gateway.ItemRepository,
	customers gateway.CustomerRepository,
	sellers gateway.SellerRepository,
	stock gateway.StockRepository,
	policy domain.StockPolicy,
) CreateSale {
	return &createSale{tx: tx, sales: sales, items: items, customers: customers, sellers: sellers, stock: stock, policy: policy}
}

func (useCase *createSale) Execute(ctx context.Context, input CreateSaleInput) (domain.Sale, error) {
	if err := input.Actor.Authorize(domain.PermissionCreateSales); err != nil {
		return domain.Sale{}, err
	}

	request, err := domain.NewSaleRequest(input.Request)
	if err != nil {
		return domain.Sale{}, err
	}

	now := time.Now().UTC()
	var sale domain.Sale
	err = useCase.tx.WithinTx(ctx, func(ctx context.Context) error {
		existing, err := useCase.sales.GetByIdempotencyKey(ctx, input.Actor.ID, request.IdempotencyKey)
		if err == nil {
			sale = existing
			return existing.Replays(request)
		}
		if !errors.Is(err, domain.ErrSaleNotFound) {
			return err
		}

		draft, err := useCase.compute(ctx, request, input.Actor, now)
		if err != nil {
			return err
		}
		if sale, err = useCase.sales.Create(ctx, draft); err != nil {
			return err
		}
		for _, line := range sale.Lines {
			if err := useCase.takeOut(ctx, sale, line); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return domain.Sale{}, domain.FromStorage(err)
	}

	return sale, nil
}

// compute reads the customer, the seller and the items the request names
// and lets the domain price the sale.
func (useCase *createSale) compute(ctx context.Context, request domain.SaleRequest, actor domain.Actor, now time.Time) (domain.Sale, error) {
	customerID := domain.FinalConsumerID
	if request.CustomerID != nil {
		customerID = *request.CustomerID
	}
	customer, err := useCase.customers.Get(ctx, customerID)
	if err != nil {
		return domain.Sale{}, referenced(err, domain.ErrCustomerNotFound, domain.ErrSaleCustomerNotFound)
	}

	var seller *domain.Seller
	if request.SellerID != nil {
		found, err := useCase.sellers.Get(ctx, *request.SellerID)
		if err != nil {
			return domain.Sale{}, referenced(err, domain.ErrSellerNotFound, domain.ErrSaleSellerNotFound)
		}
		seller = &found
	}

	items := make(map[string]domain.Item, len(request.Lines))
	for _, line := range request.Lines {
		item, err := useCase.items.Get(ctx, line.ItemID)
		if err != nil {
			return domain.Sale{}, referenced(err, domain.ErrItemNotFound, domain.ErrSaleItemNotFound)
		}
		items[item.ID] = item
	}

	return domain.NewSale(request, customer, seller, items, actor, now)
}

func (useCase *createSale) takeOut(ctx context.Context, sale domain.Sale, line domain.SaleLine) error {
	level, err := useCase.stock.GetLevel(ctx, line.ItemID)
	if err != nil {
		return err
	}
	movement, err := level.Sell(sale, line, useCase.policy)
	if err != nil {
		return err
	}
	_, err = useCase.stock.Record(ctx, movement)

	return err
}

// referenced replaces the not-found error of a record the request refers to
// with the error that blames the request instead of the URL.
func referenced(err error, notFound, replacement *domain.Error) error {
	if errors.Is(err, notFound) {
		return replacement
	}

	return err
}
