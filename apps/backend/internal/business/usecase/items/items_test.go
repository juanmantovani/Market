package items

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

func fields() domain.ItemFields {
	return domain.ItemFields{
		Code:       " YER-1 ",
		Name:       "Yerba mate 1 kg",
		CategoryID: "category-1",
		Unit:       domain.UnitPiece,
		SalePrice:  350_000,
		TaxRate:    2100,
	}
}

func storedItem() domain.Item {
	stored := domain.Item{ID: "item-1", ItemFields: fields(), CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	stored.UpdatedAt = stored.CreatedAt
	return stored
}

func assertError(t *testing.T, err error, want *domain.Error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %s", err, want.Code)
	}
}

func TestCreateItem(t *testing.T) {
	repository := &gatewayfake.ItemRepository{}
	before := time.Now().UTC()

	created, err := NewCreateItem(repository).Execute(context.Background(), CreateItemInput{Actor: manager, Fields: fields()})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if created.ID == "" || created.Code != "YER-1" || created.CreatedAt.Before(before) || created.CreatedAt.Location() != time.UTC {
		t.Fatalf("created = %+v", created)
	}
	if len(repository.Created) != 1 {
		t.Fatalf("repository received %d creates, want 1", len(repository.Created))
	}
}

func TestCreateItemRejectsBeforeStoring(t *testing.T) {
	invalid := fields()
	invalid.SalePrice = 0

	tests := []struct {
		name  string
		input CreateItemInput
		want  *domain.Error
	}{
		{name: "cashier", input: CreateItemInput{Actor: cashier, Fields: fields()}, want: domain.ErrForbidden},
		{name: "invalid price", input: CreateItemInput{Actor: admin, Fields: invalid}, want: domain.ErrItemPriceInvalid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &gatewayfake.ItemRepository{}

			_, err := NewCreateItem(repository).Execute(context.Background(), test.input)

			assertError(t, err, test.want)
			if len(repository.Created) != 0 {
				t.Fatal("repository was called")
			}
		})
	}
}

func TestCreateItemTranslatesRepositoryErrors(t *testing.T) {
	for _, repositoryErr := range []error{domain.ErrItemCodeInUse, domain.ErrItemCategoryNotFound} {
		repository := &gatewayfake.ItemRepository{CreateErr: repositoryErr}
		_, err := NewCreateItem(repository).Execute(context.Background(), CreateItemInput{Actor: admin, Fields: fields()})
		if !errors.Is(err, repositoryErr) {
			t.Fatalf("error = %v, want %v", err, repositoryErr)
		}
	}

	down := &gatewayfake.ItemRepository{CreateErr: errDown}
	_, err := NewCreateItem(down).Execute(context.Background(), CreateItemInput{Actor: admin, Fields: fields()})
	assertError(t, err, domain.ErrStorageUnavailable)
}

func TestListItems(t *testing.T) {
	page := domain.Page[domain.Item]{Items: []domain.Item{storedItem()}, Total: 1, Page: 2, Limit: 10}
	repository := &gatewayfake.ItemRepository{ListResult: page}

	got, err := NewListItems(repository).Execute(context.Background(), ListItemsInput{
		Actor: cashier, Search: " yer ", CategoryID: " category-1 ", LowStock: true, Page: 2, Limit: 10,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got.Total != 1 {
		t.Fatalf("page = %+v", got)
	}
	want := gateway.ItemFilter{Search: "yer", CategoryID: "category-1", LowStock: true, Page: domain.PageRequest{Page: 2, Limit: 10}}
	if repository.ListFilters[0] != want {
		t.Fatalf("filter = %+v, want %+v", repository.ListFilters[0], want)
	}
}

func TestListItemsErrors(t *testing.T) {
	_, err := NewListItems(&gatewayfake.ItemRepository{}).Execute(context.Background(), ListItemsInput{})
	assertError(t, err, domain.ErrForbidden)

	_, err = NewListItems(&gatewayfake.ItemRepository{}).Execute(context.Background(), ListItemsInput{Actor: admin, Limit: 500})
	assertError(t, err, domain.ErrInvalidPage)

	_, err = NewListItems(&gatewayfake.ItemRepository{ListErr: errDown}).Execute(context.Background(), ListItemsInput{Actor: admin})
	assertError(t, err, domain.ErrStorageUnavailable)
}

func TestGetItem(t *testing.T) {
	repository := &gatewayfake.ItemRepository{GetResult: storedItem()}

	got, err := NewGetItem(repository).Execute(context.Background(), GetItemInput{Actor: cashier, ID: "item-1"})
	if err != nil || got.ID != "item-1" || repository.GetIDs[0] != "item-1" {
		t.Fatalf("Execute() = %+v, %v", got, err)
	}

	_, err = NewGetItem(repository).Execute(context.Background(), GetItemInput{ID: "item-1"})
	assertError(t, err, domain.ErrForbidden)

	missing := &gatewayfake.ItemRepository{GetErr: domain.ErrItemNotFound}
	_, err = NewGetItem(missing).Execute(context.Background(), GetItemInput{Actor: admin, ID: "x"})
	assertError(t, err, domain.ErrItemNotFound)
}

func TestUpdateItem(t *testing.T) {
	repository := &gatewayfake.ItemRepository{GetResult: storedItem()}
	changed := fields()
	changed.Name = "Yerba suave"
	before := time.Now().UTC()

	updated, err := NewUpdateItem(repository).Execute(context.Background(), UpdateItemInput{Actor: admin, ID: "item-1", Fields: changed})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if updated.ID != "item-1" || updated.Name != "Yerba suave" || updated.UpdatedAt.Before(before) || !updated.CreatedAt.Equal(storedItem().CreatedAt) {
		t.Fatalf("updated = %+v", updated)
	}
	if len(repository.Updated) != 1 {
		t.Fatalf("repository received %d updates, want 1", len(repository.Updated))
	}
}

func TestUpdateItemErrors(t *testing.T) {
	invalid := fields()
	invalid.Unit = "box"

	tests := []struct {
		name       string
		repository *gatewayfake.ItemRepository
		input      UpdateItemInput
		want       *domain.Error
	}{
		{"cashier", &gatewayfake.ItemRepository{GetResult: storedItem()}, UpdateItemInput{Actor: cashier, ID: "item-1", Fields: fields()}, domain.ErrForbidden},
		{"not found", &gatewayfake.ItemRepository{GetErr: domain.ErrItemNotFound}, UpdateItemInput{Actor: admin, ID: "x", Fields: fields()}, domain.ErrItemNotFound},
		{"invalid fields", &gatewayfake.ItemRepository{GetResult: storedItem()}, UpdateItemInput{Actor: admin, ID: "item-1", Fields: invalid}, domain.ErrItemUnitInvalid},
		{"duplicate code", &gatewayfake.ItemRepository{GetResult: storedItem(), UpdateErr: domain.ErrItemCodeInUse}, UpdateItemInput{Actor: admin, ID: "item-1", Fields: fields()}, domain.ErrItemCodeInUse},
		{"storage down", &gatewayfake.ItemRepository{GetResult: storedItem(), UpdateErr: errDown}, UpdateItemInput{Actor: admin, ID: "item-1", Fields: fields()}, domain.ErrStorageUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewUpdateItem(test.repository).Execute(context.Background(), test.input)
			assertError(t, err, test.want)
		})
	}
}

func TestDeleteItem(t *testing.T) {
	repository := &gatewayfake.ItemRepository{}

	if err := NewDeleteItem(repository).Execute(context.Background(), DeleteItemInput{Actor: manager, ID: "item-1"}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(repository.DeletedIDs) != 1 || repository.DeletedIDs[0] != "item-1" {
		t.Fatalf("deleted IDs = %v", repository.DeletedIDs)
	}

	err := NewDeleteItem(&gatewayfake.ItemRepository{}).Execute(context.Background(), DeleteItemInput{Actor: cashier, ID: "item-1"})
	assertError(t, err, domain.ErrForbidden)

	err = NewDeleteItem(&gatewayfake.ItemRepository{DeleteErr: domain.ErrItemNotFound}).Execute(context.Background(), DeleteItemInput{Actor: admin, ID: "x"})
	assertError(t, err, domain.ErrItemNotFound)
}
