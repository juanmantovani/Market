package handler

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/suppliers"
	pagination "market/backend/internal/infrastructure/delivery/webapp/dto/pagination"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/suppliers"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

var handlerSupplier = domain.Supplier{
	ID: "su1",
	SupplierFields: domain.SupplierFields{
		BusinessName: "Distribuidora del Centro SA", CUIT: ptr("30701234561"), ContactName: ptr("Laura Díaz"),
		Phone: ptr("351 555-2000"), Email: ptr("pedidos@delcentro.com.ar"),
	},
	CreatedAt: handlerTime,
	UpdatedAt: handlerTime,
}

const supplierBody = `{"businessName":"Distribuidora del Centro SA","cuit":"30-70123456-1","contactName":"Laura Díaz",` +
	`"phone":"351 555-2000","email":"pedidos@delcentro.com.ar","address":null,"notes":null}`

type deleteSupplierFunc func(ctx context.Context, input suppliers.DeleteSupplierInput) error

func (fn deleteSupplierFunc) Execute(ctx context.Context, input suppliers.DeleteSupplierInput) error {
	return fn(ctx, input)
}

func assertSupplierBody(t *testing.T, got dto.SupplierResponse) {
	t.Helper()
	if got.ID != "su1" || got.BusinessName != "Distribuidora del Centro SA" || got.CUIT == nil || *got.CUIT != "30701234561" ||
		got.ContactName == nil || *got.ContactName != "Laura Díaz" || got.Address != nil {
		t.Fatalf("body = %+v", got)
	}
}

func assertSupplierFields(t *testing.T, got domain.SupplierFields) {
	t.Helper()
	if got.BusinessName != "Distribuidora del Centro SA" || got.CUIT == nil || *got.CUIT != "30-70123456-1" ||
		got.Email == nil || *got.Email != "pedidos@delcentro.com.ar" || got.Address != nil || got.Notes != nil {
		t.Fatalf("fields = %+v", got)
	}
}

func TestListSuppliersHandler(t *testing.T) {
	var received suppliers.ListSuppliersInput
	listSuppliers := useCaseFunc[suppliers.ListSuppliersInput, domain.Page[domain.Supplier]](
		func(_ context.Context, input suppliers.ListSuppliersInput) (domain.Page[domain.Supplier], error) {
			received = input
			return domain.Page[domain.Supplier]{Items: []domain.Supplier{handlerSupplier}, Total: 1, Page: 1, Limit: 25}, nil
		})

	recorder := serve(NewListSuppliersHandler(listSuppliers), categoryRequest(http.MethodGet, "/api/v1/suppliers?q=centro&page=1&limit=25", nil))

	assertStatus(t, recorder, http.StatusOK)
	got := decodeBody[pagination.PageResponse[dto.SupplierResponse]](t, recorder)
	if got.Total != 1 || len(got.Items) != 1 {
		t.Fatalf("body = %+v", got)
	}
	assertSupplierBody(t, got.Items[0])
	if received != (suppliers.ListSuppliersInput{Actor: handlerActor, Search: "centro", Page: 1, Limit: 25}) {
		t.Fatalf("input = %+v", received)
	}

	assertStatus(t, serve(NewListSuppliersHandler(listSuppliers), categoryRequest(http.MethodGet, "/api/v1/suppliers?page=x", nil)), http.StatusBadRequest)
}

func TestGetSupplierHandler(t *testing.T) {
	getSupplier := useCaseFunc[suppliers.GetSupplierInput, domain.Supplier](
		func(_ context.Context, input suppliers.GetSupplierInput) (domain.Supplier, error) {
			if input.ID != "su1" {
				return domain.Supplier{}, domain.ErrSupplierNotFound
			}
			return handlerSupplier, nil
		})

	request := categoryRequest(http.MethodGet, "/api/v1/suppliers/su1", nil)
	request.SetPathValue("id", "su1")
	recorder := serve(NewGetSupplierHandler(getSupplier), request)
	assertStatus(t, recorder, http.StatusOK)
	assertSupplierBody(t, decodeBody[dto.SupplierResponse](t, recorder))

	missing := categoryRequest(http.MethodGet, "/api/v1/suppliers/x", nil)
	missing.SetPathValue("id", "x")
	assertStatus(t, serve(NewGetSupplierHandler(getSupplier), missing), http.StatusNotFound)
}

func TestCreateSupplierHandler(t *testing.T) {
	var received suppliers.CreateSupplierInput
	createSupplier := useCaseFunc[suppliers.CreateSupplierInput, domain.Supplier](
		func(_ context.Context, input suppliers.CreateSupplierInput) (domain.Supplier, error) {
			received = input
			return handlerSupplier, nil
		})

	recorder := serve(NewCreateSupplierHandler(createSupplier), categoryRequest(http.MethodPost, "/api/v1/suppliers", strings.NewReader(supplierBody)))

	assertStatus(t, recorder, http.StatusCreated)
	assertSupplierBody(t, decodeBody[dto.SupplierResponse](t, recorder))
	assertSupplierFields(t, received.Fields)

	assertStatus(t, serve(NewCreateSupplierHandler(createSupplier), categoryRequest(http.MethodPost, "/api/v1/suppliers", strings.NewReader(`{`))), http.StatusBadRequest)
}

func TestUpdateSupplierHandler(t *testing.T) {
	var received suppliers.UpdateSupplierInput
	updateSupplier := useCaseFunc[suppliers.UpdateSupplierInput, domain.Supplier](
		func(_ context.Context, input suppliers.UpdateSupplierInput) (domain.Supplier, error) {
			received = input
			return handlerSupplier, nil
		})
	request := categoryRequest(http.MethodPut, "/api/v1/suppliers/su1", strings.NewReader(supplierBody))
	request.SetPathValue("id", "su1")

	recorder := serve(NewUpdateSupplierHandler(updateSupplier), request)

	assertStatus(t, recorder, http.StatusOK)
	assertSupplierBody(t, decodeBody[dto.SupplierResponse](t, recorder))
	assertSupplierFields(t, received.Fields)
	if received.ID != "su1" {
		t.Fatalf("id = %q", received.ID)
	}

	assertStatus(t, serve(NewUpdateSupplierHandler(updateSupplier), categoryRequest(http.MethodPut, "/api/v1/suppliers/su1", strings.NewReader(`[]`))), http.StatusBadRequest)
}

func TestDeleteSupplierHandler(t *testing.T) {
	deleteSupplier := deleteSupplierFunc(func(_ context.Context, input suppliers.DeleteSupplierInput) error {
		if input.ID != "su1" {
			return domain.ErrSupplierNotFound
		}
		return nil
	})

	request := categoryRequest(http.MethodDelete, "/api/v1/suppliers/su1", nil)
	request.SetPathValue("id", "su1")
	assertStatus(t, serve(NewDeleteSupplierHandler(deleteSupplier), request), http.StatusNoContent)

	missing := categoryRequest(http.MethodDelete, "/api/v1/suppliers/x", nil)
	missing.SetPathValue("id", "x")
	assertStatus(t, serve(NewDeleteSupplierHandler(deleteSupplier), missing), http.StatusNotFound)
}

func TestSupplierHandlersPassErrorsThrough(t *testing.T) {
	conflict := func(context.Context, suppliers.CreateSupplierInput) (domain.Supplier, error) {
		return domain.Supplier{}, domain.ErrSupplierCUITInUse
	}
	forbidden := func(context.Context, suppliers.UpdateSupplierInput) (domain.Supplier, error) {
		return domain.Supplier{}, domain.ErrForbidden
	}
	down := func(context.Context, suppliers.ListSuppliersInput) (domain.Page[domain.Supplier], error) {
		return domain.Page[domain.Supplier]{}, domain.ErrStorageUnavailable
	}

	recorder := serve(NewCreateSupplierHandler(useCaseFunc[suppliers.CreateSupplierInput, domain.Supplier](conflict)), categoryRequest(http.MethodPost, "/api/v1/suppliers", strings.NewReader(supplierBody)))
	assertStatus(t, recorder, http.StatusConflict)
	if got := decodeBody[response.ErrorResponse](t, recorder); got.Code != "supplier_cuit_in_use" {
		t.Fatalf("code = %q", got.Code)
	}

	assertStatus(t, serve(NewUpdateSupplierHandler(useCaseFunc[suppliers.UpdateSupplierInput, domain.Supplier](forbidden)), categoryRequest(http.MethodPut, "/api/v1/suppliers/su1", strings.NewReader(supplierBody))), http.StatusForbidden)
	assertStatus(t, serve(NewListSuppliersHandler(useCaseFunc[suppliers.ListSuppliersInput, domain.Page[domain.Supplier]](down)), categoryRequest(http.MethodGet, "/api/v1/suppliers", nil)), http.StatusServiceUnavailable)
}
