package dashboard

import (
	"context"
	"errors"
	"testing"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
	"market/backend/internal/business/gateway/gatewayfake"
)

var (
	cashier = domain.Actor{ID: "u-cashier", Role: domain.RoleCashier}

	errDown = errors.New("connection refused")

	from = time.Date(2026, 11, 12, 3, 0, 0, 0, time.UTC)
	to   = from.Add(24 * time.Hour)
)

func sale(total domain.Money) domain.Sale {
	return domain.Sale{
		Status: domain.SaleCompleted, Total: total,
		Payments: []domain.Payment{{Method: domain.PaymentCash, Amount: total}},
	}
}

func TestGetSummary(t *testing.T) {
	sales := &gatewayfake.SaleRepository{ListResult: domain.Page[domain.Sale]{
		Items: []domain.Sale{sale(10_000), sale(25_000)}, Total: 2, Page: 1, Limit: 100,
	}}
	low := domain.ItemStock{
		Item:  domain.Item{ID: "coca", ItemFields: domain.ItemFields{Name: "Gaseosa", MinStock: domain.WholeQuantity(5)}},
		Level: domain.StockLevel{ItemID: "coca", Quantity: domain.WholeQuantity(2)},
	}
	stock := &gatewayfake.StockRepository{ListLevelsResult: domain.Page[domain.ItemStock]{Items: []domain.ItemStock{low}, Total: 31}}

	got, err := NewGetSummary(sales, stock).Execute(context.Background(), GetSummaryInput{Actor: cashier, From: from, To: to})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got.Period.From != from || got.Period.To != to {
		t.Errorf("period = %+v", got.Period)
	}
	if got.Sales.Count != 2 || got.Sales.Total != 35_000 {
		t.Errorf("sales = %+v", got.Sales)
	}
	if got.LowStockTotal != 31 || len(got.LowStock) != 1 || got.LowStock[0].Item.ID != "coca" {
		t.Errorf("low stock = %d %+v", got.LowStockTotal, got.LowStock)
	}
	wantSales := gateway.SaleFilter{
		SaleFilter: domain.SaleFilter{From: from, To: to, Status: domain.SaleCompleted},
		Page:       domain.PageRequest{Page: 1, Limit: domain.MaxPageLimit},
	}
	if len(sales.ListFilters) != 1 || sales.ListFilters[0] != wantSales {
		t.Errorf("sale filters = %+v, want %+v", sales.ListFilters, wantSales)
	}
	if len(stock.ListLevelsFilters) != 1 || !stock.ListLevelsFilters[0].LowStock ||
		stock.ListLevelsFilters[0].Page != (domain.PageRequest{Page: 1, Limit: LowStockListSize}) {
		t.Errorf("stock filters = %+v", stock.ListLevelsFilters)
	}
}

func TestGetSummaryReadsEveryPageOfSales(t *testing.T) {
	sales := &gatewayfake.SaleRepository{ListPages: []domain.Page[domain.Sale]{
		{Items: []domain.Sale{sale(10_000), sale(20_000)}, Total: 3, Page: 1, Limit: 2},
		{Items: []domain.Sale{sale(40_000)}, Total: 3, Page: 2, Limit: 2},
	}}

	got, err := NewGetSummary(sales, &gatewayfake.StockRepository{}).Execute(context.Background(), GetSummaryInput{Actor: cashier, From: from, To: to})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got.Sales.Count != 3 || got.Sales.Total != 70_000 {
		t.Fatalf("sales = %+v", got.Sales)
	}
	if len(sales.ListFilters) != 2 || sales.ListFilters[1].Page.Page != 2 {
		t.Fatalf("filters = %+v", sales.ListFilters)
	}
}

func TestGetSummaryStopsOnAnEmptyPage(t *testing.T) {
	sales := &gatewayfake.SaleRepository{ListResult: domain.Page[domain.Sale]{Total: 5}}

	got, err := NewGetSummary(sales, &gatewayfake.StockRepository{}).Execute(context.Background(), GetSummaryInput{Actor: cashier, From: from, To: to})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got.Sales.Count != 0 || len(sales.ListFilters) != 1 {
		t.Fatalf("sales = %+v after %d reads", got.Sales, len(sales.ListFilters))
	}
}

func TestGetSummaryErrors(t *testing.T) {
	tests := []struct {
		name  string
		input GetSummaryInput
		sales *gatewayfake.SaleRepository
		stock *gatewayfake.StockRepository
		want  *domain.Error
	}{
		{name: "no actor", input: GetSummaryInput{From: from, To: to}, want: domain.ErrForbidden},
		{name: "no period", input: GetSummaryInput{Actor: cashier}, want: domain.ErrSummaryPeriodInvalid},
		{name: "reversed period", input: GetSummaryInput{Actor: cashier, From: to, To: from}, want: domain.ErrSummaryPeriodInvalid},
		{name: "sales down", input: GetSummaryInput{Actor: cashier, From: from, To: to}, sales: &gatewayfake.SaleRepository{ListErr: errDown}, want: domain.ErrStorageUnavailable},
		{name: "stock down", input: GetSummaryInput{Actor: cashier, From: from, To: to}, stock: &gatewayfake.StockRepository{ListLevelsErr: errDown}, want: domain.ErrStorageUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sales, stock := test.sales, test.stock
			if sales == nil {
				sales = &gatewayfake.SaleRepository{}
			}
			if stock == nil {
				stock = &gatewayfake.StockRepository{}
			}

			_, err := NewGetSummary(sales, stock).Execute(context.Background(), test.input)

			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %s", err, test.want.Code)
			}
		})
	}
}
