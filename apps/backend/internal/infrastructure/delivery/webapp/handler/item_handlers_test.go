package handler

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/items"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/items"
	pagination "market/backend/internal/infrastructure/delivery/webapp/dto/pagination"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

var handlerItem = domain.Item{
	ID: "i1",
	ItemFields: domain.ItemFields{
		Code: "YER-1", Barcode: ptr("7790387000123"), Name: "Yerba", CategoryID: "c1", SupplierIDs: []string{"s1", "s2"},
		Unit: domain.UnitKilogram, SalePrice: 350_000, Cost: ptr(domain.Money(210_000)), TaxRate: 2100,
		TracksStock: true, MinStock: 2500,
	},
	CreatedAt: handlerTime,
	UpdatedAt: handlerTime,
}

const itemBody = `{"code":"YER-1","barcode":"7790387000123","name":"Yerba","description":null,` +
	`"categoryId":"c1","supplierIds":["s1","s2"],"unit":"kg","salePrice":350000,"cost":210000,"taxRate":2100,"tracksStock":true,"minStock":2500}`

type deleteItemFunc func(ctx context.Context, input items.DeleteItemInput) error

func (fn deleteItemFunc) Execute(ctx context.Context, input items.DeleteItemInput) error {
	return fn(ctx, input)
}

func assertItemBody(t *testing.T, got dto.ItemResponse) {
	t.Helper()
	if got.ID != "i1" || got.Code != "YER-1" || got.Barcode == nil || *got.Barcode != "7790387000123" ||
		got.Unit != "kg" || got.SalePrice != 350_000 || got.Cost == nil || *got.Cost != 210_000 ||
		got.TaxRate != 2100 || !got.TracksStock || got.MinStock != 2500 || got.CategoryID != "c1" || len(got.SupplierIDs) != 2 {
		t.Fatalf("body = %+v", got)
	}
}

func wantFields(t *testing.T, got domain.ItemFields) {
	t.Helper()
	if got.Code != "YER-1" || got.Unit != domain.UnitKilogram || got.SalePrice != 350_000 ||
		got.Cost == nil || *got.Cost != 210_000 || got.TaxRate != 2100 || !got.TracksStock ||
		got.MinStock != 2500 || got.CategoryID != "c1" || got.Description != nil || len(got.SupplierIDs) != 2 || got.SupplierIDs[1] != "s2" {
		t.Fatalf("fields = %+v", got)
	}
}

func TestListItemsHandler(t *testing.T) {
	var received items.ListItemsInput
	listItems := useCaseFunc[items.ListItemsInput, domain.Page[domain.Item]](
		func(_ context.Context, input items.ListItemsInput) (domain.Page[domain.Item], error) {
			received = input
			return domain.Page[domain.Item]{Items: []domain.Item{handlerItem}, Total: 1, Page: 1, Limit: 25}, nil
		})

	recorder := serve(NewListItemsHandler(listItems), categoryRequest(http.MethodGet, "/api/v1/items?q=yer&categoryId=c1&page=1&limit=25", nil))

	assertStatus(t, recorder, http.StatusOK)
	got := decodeBody[pagination.PageResponse[dto.ItemResponse]](t, recorder)
	if got.Total != 1 || len(got.Items) != 1 {
		t.Fatalf("body = %+v", got)
	}
	assertItemBody(t, got.Items[0])
	want := items.ListItemsInput{Actor: handlerActor, Search: "yer", CategoryID: "c1", Page: 1, Limit: 25}
	if received != want {
		t.Fatalf("input = %+v, want %+v", received, want)
	}

	assertStatus(t, serve(NewListItemsHandler(listItems), categoryRequest(http.MethodGet, "/api/v1/items?page=x", nil)), http.StatusBadRequest)
}

func TestGetItemHandler(t *testing.T) {
	getItem := useCaseFunc[items.GetItemInput, domain.Item](
		func(_ context.Context, input items.GetItemInput) (domain.Item, error) {
			if input.ID != "i1" {
				return domain.Item{}, domain.ErrItemNotFound
			}
			return handlerItem, nil
		})

	request := categoryRequest(http.MethodGet, "/api/v1/items/i1", nil)
	request.SetPathValue("id", "i1")
	recorder := serve(NewGetItemHandler(getItem), request)
	assertStatus(t, recorder, http.StatusOK)
	assertItemBody(t, decodeBody[dto.ItemResponse](t, recorder))

	missing := categoryRequest(http.MethodGet, "/api/v1/items/x", nil)
	missing.SetPathValue("id", "x")
	assertStatus(t, serve(NewGetItemHandler(getItem), missing), http.StatusNotFound)
}

func TestCreateItemHandler(t *testing.T) {
	var received items.CreateItemInput
	createItem := useCaseFunc[items.CreateItemInput, domain.Item](
		func(_ context.Context, input items.CreateItemInput) (domain.Item, error) {
			received = input
			return handlerItem, nil
		})

	recorder := serve(NewCreateItemHandler(createItem), categoryRequest(http.MethodPost, "/api/v1/items", strings.NewReader(itemBody)))

	assertStatus(t, recorder, http.StatusCreated)
	assertItemBody(t, decodeBody[dto.ItemResponse](t, recorder))
	wantFields(t, received.Fields)
	if received.Actor != handlerActor {
		t.Fatalf("actor = %+v", received.Actor)
	}

	assertStatus(t, serve(NewCreateItemHandler(createItem), categoryRequest(http.MethodPost, "/api/v1/items", strings.NewReader(`{`))), http.StatusBadRequest)
}

func TestCreateItemHandlerWithoutCost(t *testing.T) {
	var received items.CreateItemInput
	createItem := useCaseFunc[items.CreateItemInput, domain.Item](
		func(_ context.Context, input items.CreateItemInput) (domain.Item, error) {
			received = input
			withoutCost := handlerItem
			withoutCost.Cost = nil
			withoutCost.SupplierIDs = nil
			return withoutCost, nil
		})

	body := strings.Replace(itemBody, `"cost":210000`, `"cost":null`, 1)
	recorder := serve(NewCreateItemHandler(createItem), categoryRequest(http.MethodPost, "/api/v1/items", strings.NewReader(body)))

	assertStatus(t, recorder, http.StatusCreated)
	if received.Fields.Cost != nil {
		t.Fatalf("cost = %v, want nil", *received.Fields.Cost)
	}
	if !strings.Contains(recorder.Body.String(), `"supplierIds":[]`) {
		t.Fatalf("an item without suppliers must list none, body = %s", recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"cost":null`) {
		t.Fatalf("body = %s, want cost null", recorder.Body.String())
	}
}

func TestUpdateItemHandler(t *testing.T) {
	var received items.UpdateItemInput
	updateItem := useCaseFunc[items.UpdateItemInput, domain.Item](
		func(_ context.Context, input items.UpdateItemInput) (domain.Item, error) {
			received = input
			return handlerItem, nil
		})
	request := categoryRequest(http.MethodPut, "/api/v1/items/i1", strings.NewReader(itemBody))
	request.SetPathValue("id", "i1")

	recorder := serve(NewUpdateItemHandler(updateItem), request)

	assertStatus(t, recorder, http.StatusOK)
	assertItemBody(t, decodeBody[dto.ItemResponse](t, recorder))
	wantFields(t, received.Fields)
	if received.ID != "i1" {
		t.Fatalf("id = %q", received.ID)
	}

	malformed := categoryRequest(http.MethodPut, "/api/v1/items/i1", strings.NewReader(`[]`))
	assertStatus(t, serve(NewUpdateItemHandler(updateItem), malformed), http.StatusBadRequest)
}

func TestDeleteItemHandler(t *testing.T) {
	var deleted string
	deleteItem := deleteItemFunc(func(_ context.Context, input items.DeleteItemInput) error {
		if input.ID != "i1" {
			return domain.ErrItemNotFound
		}
		deleted = input.ID
		return nil
	})

	request := categoryRequest(http.MethodDelete, "/api/v1/items/i1", nil)
	request.SetPathValue("id", "i1")
	assertStatus(t, serve(NewDeleteItemHandler(deleteItem), request), http.StatusNoContent)
	if deleted != "i1" {
		t.Fatalf("deleted = %q", deleted)
	}

	missing := categoryRequest(http.MethodDelete, "/api/v1/items/x", nil)
	missing.SetPathValue("id", "x")
	assertStatus(t, serve(NewDeleteItemHandler(deleteItem), missing), http.StatusNotFound)
}

func TestItemHandlersPassErrorsThrough(t *testing.T) {
	conflict := func(context.Context, items.CreateItemInput) (domain.Item, error) {
		return domain.Item{}, domain.ErrItemCodeInUse
	}
	invalid := func(context.Context, items.UpdateItemInput) (domain.Item, error) {
		return domain.Item{}, domain.ErrItemPriceInvalid
	}
	down := func(context.Context, items.ListItemsInput) (domain.Page[domain.Item], error) {
		return domain.Page[domain.Item]{}, domain.ErrStorageUnavailable
	}

	recorder := serve(NewCreateItemHandler(useCaseFunc[items.CreateItemInput, domain.Item](conflict)), categoryRequest(http.MethodPost, "/api/v1/items", strings.NewReader(itemBody)))
	assertStatus(t, recorder, http.StatusConflict)
	if got := decodeBody[response.ErrorResponse](t, recorder); got.Code != "item_code_in_use" {
		t.Fatalf("code = %q", got.Code)
	}

	assertStatus(t, serve(NewUpdateItemHandler(useCaseFunc[items.UpdateItemInput, domain.Item](invalid)), categoryRequest(http.MethodPut, "/api/v1/items/i1", strings.NewReader(itemBody))), http.StatusBadRequest)
	assertStatus(t, serve(NewListItemsHandler(useCaseFunc[items.ListItemsInput, domain.Page[domain.Item]](down)), categoryRequest(http.MethodGet, "/api/v1/items", nil)), http.StatusServiceUnavailable)
}
