package handler

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/sales"
	pagination "market/backend/internal/infrastructure/delivery/webapp/dto/pagination"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/sales"
)

func TestListSalesHandler(t *testing.T) {
	var received sales.ListSalesInput
	listSales := useCaseFunc[sales.ListSalesInput, domain.Page[domain.Sale]](
		func(_ context.Context, input sales.ListSalesInput) (domain.Page[domain.Sale], error) {
			received = input
			if input.Filter.Status == "pending" {
				return domain.Page[domain.Sale]{}, domain.ErrSaleStatusInvalid
			}
			return domain.Page[domain.Sale]{Items: []domain.Sale{handlerSale}, Total: 1, Page: 2, Limit: 10}, nil
		})

	recorder := serve(NewListSalesHandler(listSales), categoryRequest(http.MethodGet,
		"/api/v1/sales?from=2026-11-12T00:00:00-03:00&to=2026-11-13T03:00:00Z&customerId=cu1&sellerId=s1&status=completed&page=2&limit=10", nil))

	assertStatus(t, recorder, http.StatusOK)
	got := decodeBody[pagination.PageResponse[dto.SaleResponse]](t, recorder)
	if got.Total != 1 || got.Page != 2 || len(got.Items) != 1 || got.Items[0].Number != 12 || got.Items[0].VoidedAt != nil {
		t.Fatalf("body = %+v", got)
	}
	from := time.Date(2026, 11, 12, 3, 0, 0, 0, time.UTC)
	want := sales.ListSalesInput{
		Actor:  handlerActor,
		Filter: domain.SaleFilter{From: from, To: from.Add(24 * time.Hour), CustomerID: "cu1", SellerID: "s1", Status: domain.SaleCompleted},
		Page:   2,
		Limit:  10,
	}
	if received != want {
		t.Fatalf("input = %+v, want %+v", received, want)
	}

	serve(NewListSalesHandler(listSales), categoryRequest(http.MethodGet, "/api/v1/sales", nil))
	if !received.Filter.From.IsZero() || !received.Filter.To.IsZero() {
		t.Fatalf("a listing without dates filtered by %+v", received.Filter)
	}

	tests := []struct{ target, code string }{
		{target: "/api/v1/sales?from=2026-11-12", code: "sale_date_invalid"},
		{target: "/api/v1/sales?to=ayer", code: "sale_date_invalid"},
		{target: "/api/v1/sales?page=x", code: "invalid_page"},
		{target: "/api/v1/sales?status=pending", code: "sale_status_invalid"},
	}
	for _, test := range tests {
		recorder := serve(NewListSalesHandler(listSales), categoryRequest(http.MethodGet, test.target, nil))
		assertStatus(t, recorder, http.StatusBadRequest)
		if !strings.Contains(recorder.Body.String(), test.code) {
			t.Fatalf("GET %s = %s, want %s", test.target, recorder.Body.String(), test.code)
		}
	}
}

func TestGetSaleHandler(t *testing.T) {
	getSale := useCaseFunc[sales.GetSaleInput, domain.Sale](
		func(_ context.Context, input sales.GetSaleInput) (domain.Sale, error) {
			if input.ID != "sale-1" || input.Actor != handlerActor {
				return domain.Sale{}, domain.ErrSaleNotFound
			}
			return handlerSale, nil
		})

	request := categoryRequest(http.MethodGet, "/api/v1/sales/sale-1", nil)
	request.SetPathValue("id", "sale-1")
	recorder := serve(NewGetSaleHandler(getSale), request)

	assertStatus(t, recorder, http.StatusOK)
	if got := decodeBody[dto.SaleResponse](t, recorder); got.ID != "sale-1" || got.Total != 179_550 || len(got.Lines) != 1 {
		t.Fatalf("body = %+v", got)
	}

	missing := categoryRequest(http.MethodGet, "/api/v1/sales/x", nil)
	missing.SetPathValue("id", "x")
	assertStatus(t, serve(NewGetSaleHandler(getSale), missing), http.StatusNotFound)
}

func TestVoidSaleHandler(t *testing.T) {
	var received sales.VoidSaleInput
	voidedAt := handlerTime.Add(time.Hour)
	voidSale := useCaseFunc[sales.VoidSaleInput, domain.Sale](
		func(_ context.Context, input sales.VoidSaleInput) (domain.Sale, error) {
			received = input
			switch input.ID {
			case "voided":
				return domain.Sale{}, domain.ErrSaleAlreadyVoided
			case "sale-1":
				sale := handlerSale
				sale.Status, sale.VoidedAt, sale.VoidReason, sale.VoidedBy = domain.SaleVoided, &voidedAt, ptr("Error de carga"), ptr("u1")
				return sale, nil
			default:
				return domain.Sale{}, domain.ErrSaleNotFound
			}
		})
	voidRequest := func(id, body string) *http.Request {
		request := categoryRequest(http.MethodPost, "/api/v1/sales/"+id+"/void", strings.NewReader(body))
		request.SetPathValue("id", id)
		return request
	}

	recorder := serve(NewVoidSaleHandler(voidSale), voidRequest("sale-1", `{"reason":"Error de carga"}`))

	assertStatus(t, recorder, http.StatusOK)
	got := decodeBody[dto.SaleResponse](t, recorder)
	if got.Status != "voided" || got.VoidedAt == nil || !got.VoidedAt.Equal(voidedAt) || *got.VoidReason != "Error de carga" || *got.VoidedBy != "u1" {
		t.Fatalf("body = %+v", got)
	}
	if received != (sales.VoidSaleInput{Actor: handlerActor, ID: "sale-1", Reason: "Error de carga"}) {
		t.Fatalf("input = %+v", received)
	}

	assertStatus(t, serve(NewVoidSaleHandler(voidSale), voidRequest("voided", `{"reason":"Otra"}`)), http.StatusConflict)
	assertStatus(t, serve(NewVoidSaleHandler(voidSale), voidRequest("x", `{"reason":"Otra"}`)), http.StatusNotFound)
	assertStatus(t, serve(NewVoidSaleHandler(voidSale), voidRequest("sale-1", `{"reason":`)), http.StatusBadRequest)
}
