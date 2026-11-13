package dashboard

import (
	"context"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

// LowStockListSize is how many low stock items the summary lists; Total says
// how many there are.
const LowStockListSize = 20

type GetSummaryInput struct {
	Actor domain.Actor
	// From and To bound the period: From included, To excluded.
	From time.Time
	To   time.Time
}

type Summary struct {
	Period domain.SummaryPeriod
	Sales  domain.SalesSummary
	// LowStock lists up to LowStockListSize items by name; LowStockTotal
	// counts all of them.
	LowStock      []domain.ItemStock
	LowStockTotal int
}

// GetSummary adds up the completed sales of a period and lists the items
// whose stock reached their minimum.
type GetSummary interface {
	Execute(ctx context.Context, input GetSummaryInput) (Summary, error)
}

type getSummary struct {
	sales gateway.SaleRepository
	stock gateway.StockRepository
}

func NewGetSummary(sales gateway.SaleRepository, stock gateway.StockRepository) GetSummary {
	return &getSummary{sales: sales, stock: stock}
}

func (useCase *getSummary) Execute(ctx context.Context, input GetSummaryInput) (Summary, error) {
	if err := input.Actor.Authorize(domain.PermissionViewDashboard); err != nil {
		return Summary{}, err
	}
	period, err := domain.NewSummaryPeriod(input.From, input.To)
	if err != nil {
		return Summary{}, err
	}

	sales, err := useCase.completedSales(ctx, period)
	if err != nil {
		return Summary{}, domain.FromStorage(err)
	}
	lowStock, err := useCase.stock.ListLevels(ctx, gateway.StockLevelFilter{
		LowStock: true,
		Page:     domain.PageRequest{Page: 1, Limit: LowStockListSize},
	})
	if err != nil {
		return Summary{}, domain.FromStorage(err)
	}

	return Summary{
		Period:        period,
		Sales:         domain.NewSalesSummary(sales),
		LowStock:      lowStock.Items,
		LowStockTotal: lowStock.Total,
	}, nil
}

func (useCase *getSummary) completedSales(ctx context.Context, period domain.SummaryPeriod) ([]domain.Sale, error) {
	filter := gateway.SaleFilter{
		SaleFilter: domain.SaleFilter{From: period.From, To: period.To, Status: domain.SaleCompleted},
		Page:       domain.PageRequest{Page: 1, Limit: domain.MaxPageLimit},
	}

	var sales []domain.Sale
	for {
		page, err := useCase.sales.List(ctx, filter)
		if err != nil {
			return nil, err
		}
		sales = append(sales, page.Items...)
		if len(page.Items) == 0 || len(sales) >= page.Total {
			return sales, nil
		}
		filter.Page.Page++
	}
}
