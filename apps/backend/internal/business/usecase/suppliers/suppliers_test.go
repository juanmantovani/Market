package suppliers

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

func ptr(value string) *string {
	return &value
}

func fields() domain.SupplierFields {
	return domain.SupplierFields{
		BusinessName: " Distribuidora del Centro SA ",
		CUIT:         ptr("30-70123456-1"),
	}
}

func storedSupplier() domain.Supplier {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return domain.Supplier{ID: "supplier-1", SupplierFields: fields(), CreatedAt: created, UpdatedAt: created}
}

func assertError(t *testing.T, err error, want *domain.Error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %s", err, want.Code)
	}
}

func TestCreateSupplier(t *testing.T) {
	repository := &gatewayfake.SupplierRepository{}
	before := time.Now().UTC()

	created, err := NewCreateSupplier(repository).Execute(context.Background(), CreateSupplierInput{Actor: manager, Fields: fields()})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if created.ID == "" || created.BusinessName != "Distribuidora del Centro SA" || *created.CUIT != "30701234561" || created.CreatedAt.Before(before) {
		t.Fatalf("created = %+v", created)
	}
}

func TestCreateSupplierErrors(t *testing.T) {
	invalid := fields()
	invalid.CUIT = ptr("30-1")

	_, err := NewCreateSupplier(&gatewayfake.SupplierRepository{}).Execute(context.Background(), CreateSupplierInput{Actor: cashier, Fields: fields()})
	assertError(t, err, domain.ErrForbidden)

	repository := &gatewayfake.SupplierRepository{}
	_, err = NewCreateSupplier(repository).Execute(context.Background(), CreateSupplierInput{Actor: admin, Fields: invalid})
	assertError(t, err, domain.ErrSupplierCUITInvalid)
	if len(repository.Created) != 0 {
		t.Fatal("an invalid supplier reached the repository")
	}

	_, err = NewCreateSupplier(&gatewayfake.SupplierRepository{CreateErr: domain.ErrSupplierCUITInUse}).Execute(context.Background(), CreateSupplierInput{Actor: admin, Fields: fields()})
	assertError(t, err, domain.ErrSupplierCUITInUse)

	_, err = NewCreateSupplier(&gatewayfake.SupplierRepository{CreateErr: errDown}).Execute(context.Background(), CreateSupplierInput{Actor: admin, Fields: fields()})
	assertError(t, err, domain.ErrStorageUnavailable)
}

func TestListSuppliers(t *testing.T) {
	page := domain.Page[domain.Supplier]{Items: []domain.Supplier{storedSupplier()}, Total: 1, Page: 1, Limit: 25}
	repository := &gatewayfake.SupplierRepository{ListResult: page}

	got, err := NewListSuppliers(repository).Execute(context.Background(), ListSuppliersInput{Actor: manager, Search: " centro "})
	if err != nil || got.Total != 1 {
		t.Fatalf("Execute() = %+v, %v", got, err)
	}
	want := gateway.SupplierFilter{Search: "centro", Page: domain.PageRequest{Page: 1, Limit: domain.DefaultPageLimit}}
	if repository.ListFilters[0] != want {
		t.Fatalf("filter = %+v, want %+v", repository.ListFilters[0], want)
	}

	_, err = NewListSuppliers(repository).Execute(context.Background(), ListSuppliersInput{Actor: cashier})
	assertError(t, err, domain.ErrForbidden)
	_, err = NewListSuppliers(repository).Execute(context.Background(), ListSuppliersInput{Actor: admin, Page: -1})
	assertError(t, err, domain.ErrInvalidPage)
	_, err = NewListSuppliers(&gatewayfake.SupplierRepository{ListErr: errDown}).Execute(context.Background(), ListSuppliersInput{Actor: admin})
	assertError(t, err, domain.ErrStorageUnavailable)
}

func TestGetSupplier(t *testing.T) {
	repository := &gatewayfake.SupplierRepository{GetResult: storedSupplier()}

	got, err := NewGetSupplier(repository).Execute(context.Background(), GetSupplierInput{Actor: manager, ID: "supplier-1"})
	if err != nil || got.ID != "supplier-1" {
		t.Fatalf("Execute() = %+v, %v", got, err)
	}

	_, err = NewGetSupplier(repository).Execute(context.Background(), GetSupplierInput{Actor: cashier, ID: "supplier-1"})
	assertError(t, err, domain.ErrForbidden)
	_, err = NewGetSupplier(&gatewayfake.SupplierRepository{GetErr: domain.ErrSupplierNotFound}).Execute(context.Background(), GetSupplierInput{Actor: admin, ID: "x"})
	assertError(t, err, domain.ErrSupplierNotFound)
}

func TestUpdateSupplier(t *testing.T) {
	repository := &gatewayfake.SupplierRepository{GetResult: storedSupplier()}
	changed := fields()
	changed.BusinessName = "Distribuidora del Centro"
	changed.CUIT = nil

	updated, err := NewUpdateSupplier(repository).Execute(context.Background(), UpdateSupplierInput{Actor: manager, ID: "supplier-1", Fields: changed})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if updated.ID != "supplier-1" || updated.BusinessName != "Distribuidora del Centro" || updated.CUIT != nil || len(repository.Updated) != 1 {
		t.Fatalf("updated = %+v", updated)
	}
}

func TestUpdateSupplierErrors(t *testing.T) {
	invalid := fields()
	invalid.BusinessName = ""

	tests := []struct {
		name       string
		repository *gatewayfake.SupplierRepository
		input      UpdateSupplierInput
		want       *domain.Error
	}{
		{"cashier cannot edit", &gatewayfake.SupplierRepository{GetResult: storedSupplier()}, UpdateSupplierInput{Actor: cashier, ID: "supplier-1", Fields: fields()}, domain.ErrForbidden},
		{"not found", &gatewayfake.SupplierRepository{GetErr: domain.ErrSupplierNotFound}, UpdateSupplierInput{Actor: admin, ID: "x", Fields: fields()}, domain.ErrSupplierNotFound},
		{"invalid fields", &gatewayfake.SupplierRepository{GetResult: storedSupplier()}, UpdateSupplierInput{Actor: admin, ID: "supplier-1", Fields: invalid}, domain.ErrSupplierBusinessNameRequired},
		{"duplicate CUIT", &gatewayfake.SupplierRepository{GetResult: storedSupplier(), UpdateErr: domain.ErrSupplierCUITInUse}, UpdateSupplierInput{Actor: admin, ID: "supplier-1", Fields: fields()}, domain.ErrSupplierCUITInUse},
		{"storage down", &gatewayfake.SupplierRepository{GetResult: storedSupplier(), UpdateErr: errDown}, UpdateSupplierInput{Actor: admin, ID: "supplier-1", Fields: fields()}, domain.ErrStorageUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewUpdateSupplier(test.repository).Execute(context.Background(), test.input)
			assertError(t, err, test.want)
		})
	}
}

func TestDeleteSupplier(t *testing.T) {
	repository := &gatewayfake.SupplierRepository{}

	if err := NewDeleteSupplier(repository).Execute(context.Background(), DeleteSupplierInput{Actor: admin, ID: "supplier-1"}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(repository.DeletedIDs) != 1 || repository.DeletedIDs[0] != "supplier-1" {
		t.Fatalf("deleted IDs = %v", repository.DeletedIDs)
	}

	err := NewDeleteSupplier(&gatewayfake.SupplierRepository{}).Execute(context.Background(), DeleteSupplierInput{Actor: cashier, ID: "supplier-1"})
	assertError(t, err, domain.ErrForbidden)
	err = NewDeleteSupplier(&gatewayfake.SupplierRepository{DeleteErr: domain.ErrSupplierNotFound}).Execute(context.Background(), DeleteSupplierInput{Actor: admin, ID: "x"})
	assertError(t, err, domain.ErrSupplierNotFound)
}
