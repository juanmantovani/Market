package handler

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/customers"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/customers"
	pagination "market/backend/internal/infrastructure/delivery/webapp/dto/pagination"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

var handlerCustomer = domain.Customer{
	ID: "cu1",
	CustomerFields: domain.CustomerFields{
		Name: "La Esquina SRL", DocumentType: domain.DocumentCUIT, DocumentNumber: "30712345678",
		TaxCondition: domain.TaxRegistered, Email: ptr("compras@laesquina.com.ar"), Phone: ptr("351 555-1234"),
	},
	CreatedAt: handlerTime,
	UpdatedAt: handlerTime,
}

const customerBody = `{"name":"La Esquina SRL","documentType":"cuit","documentNumber":"30-71234567-8",` +
	`"taxCondition":"registered","email":"compras@laesquina.com.ar","phone":"351 555-1234","address":null,"notes":null}`

type deleteCustomerFunc func(ctx context.Context, input customers.DeleteCustomerInput) error

func (fn deleteCustomerFunc) Execute(ctx context.Context, input customers.DeleteCustomerInput) error {
	return fn(ctx, input)
}

func assertCustomerBody(t *testing.T, got dto.CustomerResponse) {
	t.Helper()
	if got.ID != "cu1" || got.Name != "La Esquina SRL" || got.DocumentType != "cuit" || got.DocumentNumber != "30712345678" ||
		got.TaxCondition != "registered" || got.Email == nil || *got.Email != "compras@laesquina.com.ar" || got.Address != nil {
		t.Fatalf("body = %+v", got)
	}
}

func assertCustomerFields(t *testing.T, got domain.CustomerFields) {
	t.Helper()
	if got.Name != "La Esquina SRL" || got.DocumentType != domain.DocumentCUIT || got.DocumentNumber != "30-71234567-8" ||
		got.TaxCondition != domain.TaxRegistered || got.Phone == nil || *got.Phone != "351 555-1234" || got.Address != nil {
		t.Fatalf("fields = %+v", got)
	}
}

func TestListCustomersHandler(t *testing.T) {
	var received customers.ListCustomersInput
	listCustomers := useCaseFunc[customers.ListCustomersInput, domain.Page[domain.Customer]](
		func(_ context.Context, input customers.ListCustomersInput) (domain.Page[domain.Customer], error) {
			received = input
			return domain.Page[domain.Customer]{Items: []domain.Customer{handlerCustomer}, Total: 1, Page: 1, Limit: 25}, nil
		})

	recorder := serve(NewListCustomersHandler(listCustomers), categoryRequest(http.MethodGet, "/api/v1/customers?q=esquina&page=1&limit=25", nil))

	assertStatus(t, recorder, http.StatusOK)
	got := decodeBody[pagination.PageResponse[dto.CustomerResponse]](t, recorder)
	if got.Total != 1 || len(got.Items) != 1 {
		t.Fatalf("body = %+v", got)
	}
	assertCustomerBody(t, got.Items[0])
	if received != (customers.ListCustomersInput{Actor: handlerActor, Search: "esquina", Page: 1, Limit: 25}) {
		t.Fatalf("input = %+v", received)
	}

	assertStatus(t, serve(NewListCustomersHandler(listCustomers), categoryRequest(http.MethodGet, "/api/v1/customers?limit=x", nil)), http.StatusBadRequest)
}

func TestGetCustomerHandler(t *testing.T) {
	getCustomer := useCaseFunc[customers.GetCustomerInput, domain.Customer](
		func(_ context.Context, input customers.GetCustomerInput) (domain.Customer, error) {
			if input.ID != "cu1" {
				return domain.Customer{}, domain.ErrCustomerNotFound
			}
			return handlerCustomer, nil
		})

	request := categoryRequest(http.MethodGet, "/api/v1/customers/cu1", nil)
	request.SetPathValue("id", "cu1")
	recorder := serve(NewGetCustomerHandler(getCustomer), request)
	assertStatus(t, recorder, http.StatusOK)
	assertCustomerBody(t, decodeBody[dto.CustomerResponse](t, recorder))

	missing := categoryRequest(http.MethodGet, "/api/v1/customers/x", nil)
	missing.SetPathValue("id", "x")
	assertStatus(t, serve(NewGetCustomerHandler(getCustomer), missing), http.StatusNotFound)
}

func TestCreateCustomerHandler(t *testing.T) {
	var received customers.CreateCustomerInput
	createCustomer := useCaseFunc[customers.CreateCustomerInput, domain.Customer](
		func(_ context.Context, input customers.CreateCustomerInput) (domain.Customer, error) {
			received = input
			return handlerCustomer, nil
		})

	recorder := serve(NewCreateCustomerHandler(createCustomer), categoryRequest(http.MethodPost, "/api/v1/customers", strings.NewReader(customerBody)))

	assertStatus(t, recorder, http.StatusCreated)
	assertCustomerBody(t, decodeBody[dto.CustomerResponse](t, recorder))
	assertCustomerFields(t, received.Fields)

	assertStatus(t, serve(NewCreateCustomerHandler(createCustomer), categoryRequest(http.MethodPost, "/api/v1/customers", strings.NewReader(`{`))), http.StatusBadRequest)
}

func TestUpdateCustomerHandler(t *testing.T) {
	var received customers.UpdateCustomerInput
	updateCustomer := useCaseFunc[customers.UpdateCustomerInput, domain.Customer](
		func(_ context.Context, input customers.UpdateCustomerInput) (domain.Customer, error) {
			received = input
			return handlerCustomer, nil
		})
	request := categoryRequest(http.MethodPut, "/api/v1/customers/cu1", strings.NewReader(customerBody))
	request.SetPathValue("id", "cu1")

	recorder := serve(NewUpdateCustomerHandler(updateCustomer), request)

	assertStatus(t, recorder, http.StatusOK)
	assertCustomerBody(t, decodeBody[dto.CustomerResponse](t, recorder))
	assertCustomerFields(t, received.Fields)
	if received.ID != "cu1" {
		t.Fatalf("id = %q", received.ID)
	}

	assertStatus(t, serve(NewUpdateCustomerHandler(updateCustomer), categoryRequest(http.MethodPut, "/api/v1/customers/cu1", strings.NewReader(`[]`))), http.StatusBadRequest)
}

func TestDeleteCustomerHandler(t *testing.T) {
	deleteCustomer := deleteCustomerFunc(func(_ context.Context, input customers.DeleteCustomerInput) error {
		if input.ID != "cu1" {
			return domain.ErrCustomerNotFound
		}
		return nil
	})

	request := categoryRequest(http.MethodDelete, "/api/v1/customers/cu1", nil)
	request.SetPathValue("id", "cu1")
	assertStatus(t, serve(NewDeleteCustomerHandler(deleteCustomer), request), http.StatusNoContent)

	missing := categoryRequest(http.MethodDelete, "/api/v1/customers/x", nil)
	missing.SetPathValue("id", "x")
	assertStatus(t, serve(NewDeleteCustomerHandler(deleteCustomer), missing), http.StatusNotFound)
}

func TestCustomerHandlersPassErrorsThrough(t *testing.T) {
	conflict := func(context.Context, customers.CreateCustomerInput) (domain.Customer, error) {
		return domain.Customer{}, domain.ErrCustomerDocumentInUse
	}
	forbidden := func(context.Context, customers.UpdateCustomerInput) (domain.Customer, error) {
		return domain.Customer{}, domain.ErrForbidden
	}
	down := func(context.Context, customers.ListCustomersInput) (domain.Page[domain.Customer], error) {
		return domain.Page[domain.Customer]{}, domain.ErrStorageUnavailable
	}

	recorder := serve(NewCreateCustomerHandler(useCaseFunc[customers.CreateCustomerInput, domain.Customer](conflict)), categoryRequest(http.MethodPost, "/api/v1/customers", strings.NewReader(customerBody)))
	assertStatus(t, recorder, http.StatusConflict)
	if got := decodeBody[response.ErrorResponse](t, recorder); got.Code != "customer_document_in_use" {
		t.Fatalf("code = %q", got.Code)
	}

	assertStatus(t, serve(NewUpdateCustomerHandler(useCaseFunc[customers.UpdateCustomerInput, domain.Customer](forbidden)), categoryRequest(http.MethodPut, "/api/v1/customers/cu1", strings.NewReader(customerBody))), http.StatusForbidden)
	assertStatus(t, serve(NewListCustomersHandler(useCaseFunc[customers.ListCustomersInput, domain.Page[domain.Customer]](down)), categoryRequest(http.MethodGet, "/api/v1/customers", nil)), http.StatusServiceUnavailable)
}

func TestCustomerHandlersMarkTheSystemCustomer(t *testing.T) {
	getCustomer := useCaseFunc[customers.GetCustomerInput, domain.Customer](
		func(context.Context, customers.GetCustomerInput) (domain.Customer, error) {
			return domain.FinalConsumer(handlerTime), nil
		})

	request := categoryRequest(http.MethodGet, "/api/v1/customers/"+domain.FinalConsumerID, nil)
	request.SetPathValue("id", domain.FinalConsumerID)
	recorder := serve(NewGetCustomerHandler(getCustomer), request)

	assertStatus(t, recorder, http.StatusOK)
	if got := decodeBody[dto.CustomerResponse](t, recorder); !got.System || got.Name != "Consumidor Final" || got.ID != domain.FinalConsumerID {
		t.Fatalf("body = %+v", got)
	}
}
