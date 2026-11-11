package customers

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
	admin   = domain.Actor{ID: "u-admin", Role: domain.RoleAdmin}
	manager = domain.Actor{ID: "u-manager", Role: domain.RoleManager}
	cashier = domain.Actor{ID: "u-cashier", Role: domain.RoleCashier}

	errDown = errors.New("connection refused")
)

func fields() domain.CustomerFields {
	return domain.CustomerFields{
		Name:           " Ana Pérez ",
		DocumentType:   domain.DocumentDNI,
		DocumentNumber: "12.345.678",
		TaxCondition:   domain.TaxFinalConsumer,
	}
}

func storedCustomer() domain.Customer {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return domain.Customer{ID: "customer-1", CustomerFields: fields(), CreatedAt: created, UpdatedAt: created}
}

func assertError(t *testing.T, err error, want *domain.Error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %s", err, want.Code)
	}
}

func TestCreateCustomerIsAllowedToCashiers(t *testing.T) {
	repository := &gatewayfake.CustomerRepository{}
	before := time.Now().UTC()

	created, err := NewCreateCustomer(repository).Execute(context.Background(), CreateCustomerInput{Actor: cashier, Fields: fields()})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if created.ID == "" || created.Name != "Ana Pérez" || created.DocumentNumber != "12345678" || created.CreatedAt.Before(before) {
		t.Fatalf("created = %+v", created)
	}
}

func TestCreateCustomerErrors(t *testing.T) {
	invalid := fields()
	invalid.DocumentNumber = "12"

	_, err := NewCreateCustomer(&gatewayfake.CustomerRepository{}).Execute(context.Background(), CreateCustomerInput{Fields: fields()})
	assertError(t, err, domain.ErrForbidden)

	repository := &gatewayfake.CustomerRepository{}
	_, err = NewCreateCustomer(repository).Execute(context.Background(), CreateCustomerInput{Actor: admin, Fields: invalid})
	assertError(t, err, domain.ErrCustomerDocumentInvalid)
	if len(repository.Created) != 0 {
		t.Fatal("an invalid customer reached the repository")
	}

	_, err = NewCreateCustomer(&gatewayfake.CustomerRepository{CreateErr: domain.ErrCustomerDocumentInUse}).Execute(context.Background(), CreateCustomerInput{Actor: admin, Fields: fields()})
	assertError(t, err, domain.ErrCustomerDocumentInUse)

	_, err = NewCreateCustomer(&gatewayfake.CustomerRepository{CreateErr: errDown}).Execute(context.Background(), CreateCustomerInput{Actor: admin, Fields: fields()})
	assertError(t, err, domain.ErrStorageUnavailable)
}

func TestListCustomers(t *testing.T) {
	page := domain.Page[domain.Customer]{Items: []domain.Customer{storedCustomer()}, Total: 1, Page: 1, Limit: 25}
	repository := &gatewayfake.CustomerRepository{ListResult: page}

	got, err := NewListCustomers(repository).Execute(context.Background(), ListCustomersInput{Actor: cashier, Search: " ana "})
	if err != nil || got.Total != 1 {
		t.Fatalf("Execute() = %+v, %v", got, err)
	}
	want := gateway.CustomerFilter{Search: "ana", Page: domain.PageRequest{Page: 1, Limit: domain.DefaultPageLimit}}
	if repository.ListFilters[0] != want {
		t.Fatalf("filter = %+v, want %+v", repository.ListFilters[0], want)
	}

	_, err = NewListCustomers(repository).Execute(context.Background(), ListCustomersInput{})
	assertError(t, err, domain.ErrForbidden)
	_, err = NewListCustomers(repository).Execute(context.Background(), ListCustomersInput{Actor: admin, Page: -1})
	assertError(t, err, domain.ErrInvalidPage)
	_, err = NewListCustomers(&gatewayfake.CustomerRepository{ListErr: errDown}).Execute(context.Background(), ListCustomersInput{Actor: admin})
	assertError(t, err, domain.ErrStorageUnavailable)
}

func TestGetCustomer(t *testing.T) {
	repository := &gatewayfake.CustomerRepository{GetResult: storedCustomer()}

	got, err := NewGetCustomer(repository).Execute(context.Background(), GetCustomerInput{Actor: cashier, ID: "customer-1"})
	if err != nil || got.ID != "customer-1" {
		t.Fatalf("Execute() = %+v, %v", got, err)
	}

	_, err = NewGetCustomer(repository).Execute(context.Background(), GetCustomerInput{ID: "customer-1"})
	assertError(t, err, domain.ErrForbidden)
	_, err = NewGetCustomer(&gatewayfake.CustomerRepository{GetErr: domain.ErrCustomerNotFound}).Execute(context.Background(), GetCustomerInput{Actor: admin, ID: "x"})
	assertError(t, err, domain.ErrCustomerNotFound)
}

func TestUpdateCustomer(t *testing.T) {
	repository := &gatewayfake.CustomerRepository{GetResult: storedCustomer()}
	changed := fields()
	changed.Name = "Ana María Pérez"

	updated, err := NewUpdateCustomer(repository).Execute(context.Background(), UpdateCustomerInput{Actor: manager, ID: "customer-1", Fields: changed})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if updated.ID != "customer-1" || updated.Name != "Ana María Pérez" || len(repository.Updated) != 1 {
		t.Fatalf("updated = %+v", updated)
	}
}

func TestUpdateCustomerErrors(t *testing.T) {
	invalid := fields()
	invalid.TaxCondition = domain.TaxRegistered

	tests := []struct {
		name       string
		repository *gatewayfake.CustomerRepository
		input      UpdateCustomerInput
		want       *domain.Error
	}{
		{"cashier cannot edit", &gatewayfake.CustomerRepository{GetResult: storedCustomer()}, UpdateCustomerInput{Actor: cashier, ID: "customer-1", Fields: fields()}, domain.ErrForbidden},
		{"not found", &gatewayfake.CustomerRepository{GetErr: domain.ErrCustomerNotFound}, UpdateCustomerInput{Actor: admin, ID: "x", Fields: fields()}, domain.ErrCustomerNotFound},
		{"invalid fields", &gatewayfake.CustomerRepository{GetResult: storedCustomer()}, UpdateCustomerInput{Actor: admin, ID: "customer-1", Fields: invalid}, domain.ErrCustomerTaxConditionRequiresCUIT},
		{"duplicate document", &gatewayfake.CustomerRepository{GetResult: storedCustomer(), UpdateErr: domain.ErrCustomerDocumentInUse}, UpdateCustomerInput{Actor: admin, ID: "customer-1", Fields: fields()}, domain.ErrCustomerDocumentInUse},
		{"storage down", &gatewayfake.CustomerRepository{GetResult: storedCustomer(), UpdateErr: errDown}, UpdateCustomerInput{Actor: admin, ID: "customer-1", Fields: fields()}, domain.ErrStorageUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewUpdateCustomer(test.repository).Execute(context.Background(), test.input)
			assertError(t, err, test.want)
		})
	}
}

func TestDeleteCustomer(t *testing.T) {
	repository := &gatewayfake.CustomerRepository{}

	if err := NewDeleteCustomer(repository).Execute(context.Background(), DeleteCustomerInput{Actor: admin, ID: "customer-1"}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(repository.DeletedIDs) != 1 || repository.DeletedIDs[0] != "customer-1" {
		t.Fatalf("deleted IDs = %v", repository.DeletedIDs)
	}

	err := NewDeleteCustomer(&gatewayfake.CustomerRepository{}).Execute(context.Background(), DeleteCustomerInput{Actor: cashier, ID: "customer-1"})
	assertError(t, err, domain.ErrForbidden)
	err = NewDeleteCustomer(&gatewayfake.CustomerRepository{DeleteErr: domain.ErrCustomerNotFound}).Execute(context.Background(), DeleteCustomerInput{Actor: admin, ID: "x"})
	assertError(t, err, domain.ErrCustomerNotFound)
}
