package handler

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/sales"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/sales"
)

var handlerSale = domain.Sale{
	ID: "sale-1", Number: 12, Status: domain.SaleCompleted,
	Customer: domain.SaleCustomer{ID: "cu1", Name: "Ana Pérez", DocumentType: domain.DocumentDNI, DocumentNumber: "28456789", TaxCondition: domain.TaxFinalConsumer},
	Seller:   &domain.SaleSeller{ID: "s1", Name: "Juan Pérez", Code: "JP"},
	Lines: []domain.SaleLine{{
		ItemID: "i1", Code: "PAN-1", Name: "Pan francés", Unit: domain.UnitKilogram, TaxRate: 1050,
		Quantity: 750, UnitPrice: 280_000, DiscountPercent: 1_000, DiscountAmount: 21_000, LineTotal: 189_000,
	}},
	Subtotal: 210_000, DiscountTotal: 21_000, Total: 189_000,
	Payments: []domain.Payment{{Method: domain.PaymentDebitCard, Amount: 100_000, Reference: ptr("Cupón 1")}, {Method: domain.PaymentCash, Amount: 100_000}},
	Change:   11_000, ActorID: "u1", IdempotencyKey: "key-1", CreatedAt: handlerTime,
}

func saleRequest(body, key string) *http.Request {
	request := categoryRequest(http.MethodPost, "/api/v1/sales", strings.NewReader(body))
	if key != "" {
		request.Header.Set("Idempotency-Key", key)
	}
	return request
}

func TestCreateSaleHandler(t *testing.T) {
	var received sales.CreateSaleInput
	createSale := useCaseFunc[sales.CreateSaleInput, domain.Sale](
		func(_ context.Context, input sales.CreateSaleInput) (domain.Sale, error) {
			received = input
			return handlerSale, nil
		})
	body := `{"customerId":"cu1","sellerId":"s1","lines":[{"itemId":"i1","quantity":750,"discountPercent":1000}],` +
		`"payments":[{"method":"debit_card","amount":100000,"reference":"Cupón 1"},{"method":"cash","amount":100000,"reference":null}],"expectedTotal":189000}`

	recorder := serve(NewCreateSaleHandler(createSale), saleRequest(body, "key-1"))

	assertStatus(t, recorder, http.StatusCreated)
	got := decodeBody[dto.SaleResponse](t, recorder)
	if got.ID != "sale-1" || got.Number != 12 || got.Status != "completed" || got.Subtotal != 210_000 || got.DiscountTotal != 21_000 ||
		got.Total != 189_000 || got.Change != 11_000 || got.ActorID != "u1" || !got.CreatedAt.Equal(handlerTime) {
		t.Fatalf("body = %+v", got)
	}
	if got.Customer != (dto.SaleCustomerResponse{ID: "cu1", Name: "Ana Pérez", DocumentType: "dni", DocumentNumber: "28456789", TaxCondition: "final_consumer"}) ||
		got.Seller == nil || *got.Seller != (dto.SaleSellerResponse{ID: "s1", Name: "Juan Pérez", Code: "JP"}) {
		t.Fatalf("parties = %+v, %+v", got.Customer, got.Seller)
	}
	wantLine := dto.SaleLineResponse{ItemID: "i1", Code: "PAN-1", Name: "Pan francés", Unit: "kg", TaxRate: 1050, Quantity: 750, UnitPrice: 280_000, DiscountPercent: 1_000, DiscountAmount: 21_000, LineTotal: 189_000}
	if len(got.Lines) != 1 || got.Lines[0] != wantLine {
		t.Fatalf("lines = %+v", got.Lines)
	}
	if len(got.Payments) != 2 || got.Payments[0].Method != "debit_card" || got.Payments[0].Amount != 100_000 || *got.Payments[0].Reference != "Cupón 1" ||
		got.Payments[1].Method != "cash" || got.Payments[1].Reference != nil {
		t.Fatalf("payments = %+v", got.Payments)
	}

	request := received.Request
	if received.Actor != handlerActor || request.IdempotencyKey != "key-1" || *request.CustomerID != "cu1" || *request.SellerID != "s1" || request.ExpectedTotal != 189_000 {
		t.Fatalf("input = %+v", received)
	}
	if len(request.Lines) != 1 || request.Lines[0] != (domain.SaleRequestLine{ItemID: "i1", Quantity: 750, DiscountPercent: 1_000}) {
		t.Fatalf("lines = %+v", request.Lines)
	}
	if len(request.Payments) != 2 || request.Payments[0].Method != domain.PaymentDebitCard || request.Payments[0].Amount != 100_000 ||
		*request.Payments[0].Reference != "Cupón 1" || request.Payments[1].Reference != nil {
		t.Fatalf("payments = %+v", request.Payments)
	}
}

func TestCreateSaleHandlerAnswersASaleWithoutSeller(t *testing.T) {
	createSale := useCaseFunc[sales.CreateSaleInput, domain.Sale](
		func(context.Context, sales.CreateSaleInput) (domain.Sale, error) {
			sale := handlerSale
			sale.Seller = nil
			return sale, nil
		})

	recorder := serve(NewCreateSaleHandler(createSale), saleRequest(`{"lines":[]}`, "key-1"))

	assertStatus(t, recorder, http.StatusCreated)
	if !strings.Contains(recorder.Body.String(), `"seller":null`) {
		t.Fatalf("body = %s, want a null seller", recorder.Body.String())
	}
}

func TestCreateSaleHandlerErrors(t *testing.T) {
	createSale := useCaseFunc[sales.CreateSaleInput, domain.Sale](
		func(_ context.Context, input sales.CreateSaleInput) (domain.Sale, error) {
			switch {
			case input.Request.IdempotencyKey == "":
				return domain.Sale{}, domain.ErrIdempotencyKeyRequired
			case input.Request.ExpectedTotal == 1:
				return domain.Sale{}, domain.ErrSaleTotalMismatch
			case input.Request.ExpectedTotal == 2:
				return domain.Sale{}, domain.InsufficientStock("Queso cremoso")
			default:
				return domain.Sale{}, domain.ErrForbidden
			}
		})
	tests := []struct {
		name, body, key, code string
		status                int
	}{
		{name: "invalid body", body: `{"lines":`, key: "key-1", code: "invalid_body", status: http.StatusBadRequest},
		{name: "no idempotency key", body: `{}`, code: "idempotency_key_required", status: http.StatusBadRequest},
		{name: "total mismatch", body: `{"expectedTotal":1}`, key: "key-1", code: "sale_total_mismatch", status: http.StatusConflict},
		{name: "insufficient stock", body: `{"expectedTotal":2}`, key: "key-1", code: "insufficient_stock", status: http.StatusConflict},
		{name: "forbidden", body: `{}`, key: "key-1", code: "forbidden", status: http.StatusForbidden},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := serve(NewCreateSaleHandler(createSale), saleRequest(test.body, test.key))

			assertStatus(t, recorder, test.status)
			if !strings.Contains(recorder.Body.String(), `"code":"`+test.code+`"`) {
				t.Fatalf("body = %s, want code %s", recorder.Body.String(), test.code)
			}
		})
	}
}
