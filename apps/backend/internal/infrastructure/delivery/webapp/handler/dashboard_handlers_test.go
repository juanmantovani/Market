package handler

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/dashboard"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/dashboard"
)

func TestGetDashboardSummaryHandler(t *testing.T) {
	var received dashboard.GetSummaryInput
	getSummary := useCaseFunc[dashboard.GetSummaryInput, dashboard.Summary](
		func(_ context.Context, input dashboard.GetSummaryInput) (dashboard.Summary, error) {
			received = input
			period, err := domain.NewSummaryPeriod(input.From, input.To)
			if err != nil {
				return dashboard.Summary{}, err
			}
			low := domain.ItemStock{
				Item:  domain.Item{ID: "coca", ItemFields: domain.ItemFields{Code: "COCA", Name: "Gaseosa", MinStock: domain.WholeQuantity(5)}},
				Level: domain.StockLevel{ItemID: "coca", Quantity: domain.WholeQuantity(2), UpdatedAt: handlerTime},
			}
			return dashboard.Summary{
				Period: period,
				Sales: domain.SalesSummary{Count: 3, Total: 330_000, ByMethod: []domain.MethodTotal{
					{Method: domain.PaymentCash, Amount: 180_000},
					{Method: domain.PaymentQR, Amount: 150_000},
				}},
				LowStock:      []domain.ItemStock{low},
				LowStockTotal: 4,
			}, nil
		})

	recorder := serve(NewGetDashboardSummaryHandler(getSummary), categoryRequest(http.MethodGet,
		"/api/v1/dashboard/summary?from=2026-11-12T00:00:00-03:00&to=2026-11-13T03:00:00Z", nil))

	assertStatus(t, recorder, http.StatusOK)
	got := decodeBody[dto.SummaryResponse](t, recorder)
	if got.Sales.Count != 3 || got.Sales.Total != 330_000 || len(got.Sales.ByMethod) != 2 ||
		got.Sales.ByMethod[1].Method != "qr" || got.Sales.ByMethod[1].Amount != 150_000 {
		t.Fatalf("sales = %+v", got.Sales)
	}
	if got.LowStock.Total != 4 || len(got.LowStock.Items) != 1 || got.LowStock.Items[0].ItemID != "coca" ||
		!got.LowStock.Items[0].LowStock || got.LowStock.Items[0].Quantity != 2_000 {
		t.Fatalf("low stock = %+v", got.LowStock)
	}
	from := time.Date(2026, 11, 12, 3, 0, 0, 0, time.UTC)
	want := dashboard.GetSummaryInput{Actor: handlerActor, From: from, To: from.Add(24 * time.Hour)}
	if received != want || !got.From.Equal(from) {
		t.Fatalf("input = %+v, want %+v", received, want)
	}
}

func TestGetDashboardSummaryHandlerErrors(t *testing.T) {
	getSummary := useCaseFunc[dashboard.GetSummaryInput, dashboard.Summary](
		func(_ context.Context, input dashboard.GetSummaryInput) (dashboard.Summary, error) {
			_, err := domain.NewSummaryPeriod(input.From, input.To)
			return dashboard.Summary{}, err
		})

	tests := []struct{ target, code string }{
		{target: "/api/v1/dashboard/summary", code: "summary_period_invalid"},
		{target: "/api/v1/dashboard/summary?from=2026-11-12T00:00:00Z", code: "summary_period_invalid"},
		{target: "/api/v1/dashboard/summary?from=hoy&to=2026-11-13T00:00:00Z", code: "sale_date_invalid"},
		{target: "/api/v1/dashboard/summary?from=2026-11-12T00:00:00Z&to=mañana", code: "sale_date_invalid"},
	}
	for _, test := range tests {
		recorder := serve(NewGetDashboardSummaryHandler(getSummary), categoryRequest(http.MethodGet, test.target, nil))
		assertStatus(t, recorder, http.StatusBadRequest)
		if !strings.Contains(recorder.Body.String(), test.code) {
			t.Fatalf("GET %s = %s, want %s", test.target, recorder.Body.String(), test.code)
		}
	}
}
