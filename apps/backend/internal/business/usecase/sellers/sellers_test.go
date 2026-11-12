package sellers

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

func fields() domain.SellerFields {
	rate := domain.Percent(250)
	return domain.SellerFields{Name: " Juan Pérez ", Code: "jp", CommissionRate: &rate}
}

func storedSeller() domain.Seller {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return domain.Seller{ID: "seller-1", SellerFields: fields(), CreatedAt: created, UpdatedAt: created}
}

func assertError(t *testing.T, err error, want *domain.Error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %s", err, want.Code)
	}
}

func TestCreateSeller(t *testing.T) {
	repository := &gatewayfake.SellerRepository{}
	before := time.Now().UTC()

	created, err := NewCreateSeller(repository).Execute(context.Background(), CreateSellerInput{Actor: manager, Fields: fields()})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if created.ID == "" || created.Name != "Juan Pérez" || created.Code != "JP" || *created.CommissionRate != 250 || created.CreatedAt.Before(before) {
		t.Fatalf("created = %+v", created)
	}
}

func TestCreateSellerErrors(t *testing.T) {
	invalid := fields()
	invalid.Code = "J P"

	_, err := NewCreateSeller(&gatewayfake.SellerRepository{}).Execute(context.Background(), CreateSellerInput{Actor: cashier, Fields: fields()})
	assertError(t, err, domain.ErrForbidden)

	repository := &gatewayfake.SellerRepository{}
	_, err = NewCreateSeller(repository).Execute(context.Background(), CreateSellerInput{Actor: admin, Fields: invalid})
	assertError(t, err, domain.ErrSellerCodeInvalid)
	if len(repository.Created) != 0 {
		t.Fatal("an invalid seller reached the repository")
	}

	_, err = NewCreateSeller(&gatewayfake.SellerRepository{CreateErr: domain.ErrSellerCodeInUse}).Execute(context.Background(), CreateSellerInput{Actor: admin, Fields: fields()})
	assertError(t, err, domain.ErrSellerCodeInUse)

	_, err = NewCreateSeller(&gatewayfake.SellerRepository{CreateErr: errDown}).Execute(context.Background(), CreateSellerInput{Actor: admin, Fields: fields()})
	assertError(t, err, domain.ErrStorageUnavailable)
}

func TestListSellersIsAllowedToCashiers(t *testing.T) {
	page := domain.Page[domain.Seller]{Items: []domain.Seller{storedSeller()}, Total: 1, Page: 1, Limit: 25}
	repository := &gatewayfake.SellerRepository{ListResult: page}

	got, err := NewListSellers(repository).Execute(context.Background(), ListSellersInput{Actor: cashier, Search: " jp "})
	if err != nil || got.Total != 1 {
		t.Fatalf("Execute() = %+v, %v", got, err)
	}
	want := gateway.SellerFilter{Search: "jp", Page: domain.PageRequest{Page: 1, Limit: domain.DefaultPageLimit}}
	if repository.ListFilters[0] != want {
		t.Fatalf("filter = %+v, want %+v", repository.ListFilters[0], want)
	}

	_, err = NewListSellers(repository).Execute(context.Background(), ListSellersInput{})
	assertError(t, err, domain.ErrForbidden)
	_, err = NewListSellers(repository).Execute(context.Background(), ListSellersInput{Actor: admin, Page: -1})
	assertError(t, err, domain.ErrInvalidPage)
	_, err = NewListSellers(&gatewayfake.SellerRepository{ListErr: errDown}).Execute(context.Background(), ListSellersInput{Actor: admin})
	assertError(t, err, domain.ErrStorageUnavailable)
}

func TestGetSeller(t *testing.T) {
	repository := &gatewayfake.SellerRepository{GetResult: storedSeller()}

	got, err := NewGetSeller(repository).Execute(context.Background(), GetSellerInput{Actor: cashier, ID: "seller-1"})
	if err != nil || got.ID != "seller-1" {
		t.Fatalf("Execute() = %+v, %v", got, err)
	}

	_, err = NewGetSeller(repository).Execute(context.Background(), GetSellerInput{ID: "seller-1"})
	assertError(t, err, domain.ErrForbidden)
	_, err = NewGetSeller(&gatewayfake.SellerRepository{GetErr: domain.ErrSellerNotFound}).Execute(context.Background(), GetSellerInput{Actor: admin, ID: "x"})
	assertError(t, err, domain.ErrSellerNotFound)
}

func TestUpdateSeller(t *testing.T) {
	repository := &gatewayfake.SellerRepository{GetResult: storedSeller()}
	changed := fields()
	changed.Name = "Juan Pablo Pérez"
	changed.CommissionRate = nil

	updated, err := NewUpdateSeller(repository).Execute(context.Background(), UpdateSellerInput{Actor: manager, ID: "seller-1", Fields: changed})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if updated.ID != "seller-1" || updated.Name != "Juan Pablo Pérez" || updated.CommissionRate != nil || len(repository.Updated) != 1 {
		t.Fatalf("updated = %+v", updated)
	}
}

func TestUpdateSellerErrors(t *testing.T) {
	invalid := fields()
	invalid.Name = ""

	tests := []struct {
		name       string
		repository *gatewayfake.SellerRepository
		input      UpdateSellerInput
		want       *domain.Error
	}{
		{"cashier cannot edit", &gatewayfake.SellerRepository{GetResult: storedSeller()}, UpdateSellerInput{Actor: cashier, ID: "seller-1", Fields: fields()}, domain.ErrForbidden},
		{"not found", &gatewayfake.SellerRepository{GetErr: domain.ErrSellerNotFound}, UpdateSellerInput{Actor: admin, ID: "x", Fields: fields()}, domain.ErrSellerNotFound},
		{"invalid fields", &gatewayfake.SellerRepository{GetResult: storedSeller()}, UpdateSellerInput{Actor: admin, ID: "seller-1", Fields: invalid}, domain.ErrSellerNameRequired},
		{"duplicate code", &gatewayfake.SellerRepository{GetResult: storedSeller(), UpdateErr: domain.ErrSellerCodeInUse}, UpdateSellerInput{Actor: admin, ID: "seller-1", Fields: fields()}, domain.ErrSellerCodeInUse},
		{"storage down", &gatewayfake.SellerRepository{GetResult: storedSeller(), UpdateErr: errDown}, UpdateSellerInput{Actor: admin, ID: "seller-1", Fields: fields()}, domain.ErrStorageUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewUpdateSeller(test.repository).Execute(context.Background(), test.input)
			assertError(t, err, test.want)
		})
	}
}

func TestDeleteSeller(t *testing.T) {
	repository := &gatewayfake.SellerRepository{}

	if err := NewDeleteSeller(repository).Execute(context.Background(), DeleteSellerInput{Actor: admin, ID: "seller-1"}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(repository.DeletedIDs) != 1 || repository.DeletedIDs[0] != "seller-1" {
		t.Fatalf("deleted IDs = %v", repository.DeletedIDs)
	}

	err := NewDeleteSeller(&gatewayfake.SellerRepository{}).Execute(context.Background(), DeleteSellerInput{Actor: cashier, ID: "seller-1"})
	assertError(t, err, domain.ErrForbidden)
	err = NewDeleteSeller(&gatewayfake.SellerRepository{DeleteErr: domain.ErrSellerNotFound}).Execute(context.Background(), DeleteSellerInput{Actor: admin, ID: "x"})
	assertError(t, err, domain.ErrSellerNotFound)
}

func TestDeleteSellerWithSales(t *testing.T) {
	err := NewDeleteSeller(&gatewayfake.SellerRepository{DeleteErr: domain.ErrSellerInUse}).Execute(context.Background(), DeleteSellerInput{Actor: admin, ID: "seller-1"})

	assertError(t, err, domain.ErrSellerInUse)
}
