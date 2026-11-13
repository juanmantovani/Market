package handler

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/sellers"
	pagination "market/backend/internal/infrastructure/delivery/webapp/dto/pagination"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/sellers"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

var handlerSeller = domain.Seller{
	ID:           "se1",
	SellerFields: domain.SellerFields{Name: "Juan Pérez", Code: "JP", CommissionRate: ptr(domain.Percent(250))},
	CreatedAt:    handlerTime,
	UpdatedAt:    handlerTime,
}

const sellerBody = `{"name":"Juan Pérez","code":"jp","commissionRate":250}`

type deleteSellerFunc func(ctx context.Context, input sellers.DeleteSellerInput) error

func (fn deleteSellerFunc) Execute(ctx context.Context, input sellers.DeleteSellerInput) error {
	return fn(ctx, input)
}

func assertSellerBody(t *testing.T, got dto.SellerResponse) {
	t.Helper()
	if got.ID != "se1" || got.Name != "Juan Pérez" || got.Code != "JP" || got.CommissionRate == nil || *got.CommissionRate != 250 {
		t.Fatalf("body = %+v", got)
	}
}

func assertSellerFields(t *testing.T, got domain.SellerFields) {
	t.Helper()
	if got.Name != "Juan Pérez" || got.Code != "jp" || got.CommissionRate == nil || *got.CommissionRate != 250 {
		t.Fatalf("fields = %+v", got)
	}
}

func TestListSellersHandler(t *testing.T) {
	var received sellers.ListSellersInput
	withoutCommission := handlerSeller
	withoutCommission.CommissionRate = nil
	listSellers := useCaseFunc[sellers.ListSellersInput, domain.Page[domain.Seller]](
		func(_ context.Context, input sellers.ListSellersInput) (domain.Page[domain.Seller], error) {
			received = input
			return domain.Page[domain.Seller]{Items: []domain.Seller{handlerSeller, withoutCommission}, Total: 2, Page: 1, Limit: 25}, nil
		})

	recorder := serve(NewListSellersHandler(listSellers), categoryRequest(http.MethodGet, "/api/v1/sellers?q=jp&page=1&limit=25", nil))

	assertStatus(t, recorder, http.StatusOK)
	got := decodeBody[pagination.PageResponse[dto.SellerResponse]](t, recorder)
	if got.Total != 2 || len(got.Items) != 2 || got.Items[1].CommissionRate != nil {
		t.Fatalf("body = %+v", got)
	}
	assertSellerBody(t, got.Items[0])
	if received != (sellers.ListSellersInput{Actor: handlerActor, Search: "jp", Page: 1, Limit: 25}) {
		t.Fatalf("input = %+v", received)
	}

	assertStatus(t, serve(NewListSellersHandler(listSellers), categoryRequest(http.MethodGet, "/api/v1/sellers?limit=x", nil)), http.StatusBadRequest)
}

func TestGetSellerHandler(t *testing.T) {
	getSeller := useCaseFunc[sellers.GetSellerInput, domain.Seller](
		func(_ context.Context, input sellers.GetSellerInput) (domain.Seller, error) {
			if input.ID != "se1" {
				return domain.Seller{}, domain.ErrSellerNotFound
			}
			return handlerSeller, nil
		})

	request := categoryRequest(http.MethodGet, "/api/v1/sellers/se1", nil)
	request.SetPathValue("id", "se1")
	recorder := serve(NewGetSellerHandler(getSeller), request)
	assertStatus(t, recorder, http.StatusOK)
	assertSellerBody(t, decodeBody[dto.SellerResponse](t, recorder))

	missing := categoryRequest(http.MethodGet, "/api/v1/sellers/x", nil)
	missing.SetPathValue("id", "x")
	assertStatus(t, serve(NewGetSellerHandler(getSeller), missing), http.StatusNotFound)
}

func TestCreateSellerHandler(t *testing.T) {
	var received sellers.CreateSellerInput
	createSeller := useCaseFunc[sellers.CreateSellerInput, domain.Seller](
		func(_ context.Context, input sellers.CreateSellerInput) (domain.Seller, error) {
			received = input
			return handlerSeller, nil
		})

	recorder := serve(NewCreateSellerHandler(createSeller), categoryRequest(http.MethodPost, "/api/v1/sellers", strings.NewReader(sellerBody)))

	assertStatus(t, recorder, http.StatusCreated)
	assertSellerBody(t, decodeBody[dto.SellerResponse](t, recorder))
	assertSellerFields(t, received.Fields)

	serve(NewCreateSellerHandler(createSeller), categoryRequest(http.MethodPost, "/api/v1/sellers", strings.NewReader(`{"name":"Ana","code":"AN","commissionRate":null}`)))
	if received.Fields.CommissionRate != nil {
		t.Fatalf("commission = %v, want none", *received.Fields.CommissionRate)
	}

	assertStatus(t, serve(NewCreateSellerHandler(createSeller), categoryRequest(http.MethodPost, "/api/v1/sellers", strings.NewReader(`{`))), http.StatusBadRequest)
}

func TestUpdateSellerHandler(t *testing.T) {
	var received sellers.UpdateSellerInput
	updateSeller := useCaseFunc[sellers.UpdateSellerInput, domain.Seller](
		func(_ context.Context, input sellers.UpdateSellerInput) (domain.Seller, error) {
			received = input
			return handlerSeller, nil
		})
	request := categoryRequest(http.MethodPut, "/api/v1/sellers/se1", strings.NewReader(sellerBody))
	request.SetPathValue("id", "se1")

	recorder := serve(NewUpdateSellerHandler(updateSeller), request)

	assertStatus(t, recorder, http.StatusOK)
	assertSellerBody(t, decodeBody[dto.SellerResponse](t, recorder))
	assertSellerFields(t, received.Fields)
	if received.ID != "se1" {
		t.Fatalf("id = %q", received.ID)
	}

	assertStatus(t, serve(NewUpdateSellerHandler(updateSeller), categoryRequest(http.MethodPut, "/api/v1/sellers/se1", strings.NewReader(`[]`))), http.StatusBadRequest)
}

func TestDeleteSellerHandler(t *testing.T) {
	deleteSeller := deleteSellerFunc(func(_ context.Context, input sellers.DeleteSellerInput) error {
		if input.ID != "se1" {
			return domain.ErrSellerNotFound
		}
		return nil
	})

	request := categoryRequest(http.MethodDelete, "/api/v1/sellers/se1", nil)
	request.SetPathValue("id", "se1")
	assertStatus(t, serve(NewDeleteSellerHandler(deleteSeller), request), http.StatusNoContent)

	missing := categoryRequest(http.MethodDelete, "/api/v1/sellers/x", nil)
	missing.SetPathValue("id", "x")
	assertStatus(t, serve(NewDeleteSellerHandler(deleteSeller), missing), http.StatusNotFound)
}

func TestSellerHandlersPassErrorsThrough(t *testing.T) {
	conflict := func(context.Context, sellers.CreateSellerInput) (domain.Seller, error) {
		return domain.Seller{}, domain.ErrSellerCodeInUse
	}
	forbidden := func(context.Context, sellers.UpdateSellerInput) (domain.Seller, error) {
		return domain.Seller{}, domain.ErrForbidden
	}
	down := func(context.Context, sellers.ListSellersInput) (domain.Page[domain.Seller], error) {
		return domain.Page[domain.Seller]{}, domain.ErrStorageUnavailable
	}

	recorder := serve(NewCreateSellerHandler(useCaseFunc[sellers.CreateSellerInput, domain.Seller](conflict)), categoryRequest(http.MethodPost, "/api/v1/sellers", strings.NewReader(sellerBody)))
	assertStatus(t, recorder, http.StatusConflict)
	if got := decodeBody[response.ErrorResponse](t, recorder); got.Code != "seller_code_in_use" {
		t.Fatalf("code = %q", got.Code)
	}

	assertStatus(t, serve(NewUpdateSellerHandler(useCaseFunc[sellers.UpdateSellerInput, domain.Seller](forbidden)), categoryRequest(http.MethodPut, "/api/v1/sellers/se1", strings.NewReader(sellerBody))), http.StatusForbidden)
	assertStatus(t, serve(NewListSellersHandler(useCaseFunc[sellers.ListSellersInput, domain.Page[domain.Seller]](down)), categoryRequest(http.MethodGet, "/api/v1/sellers", nil)), http.StatusServiceUnavailable)
}
