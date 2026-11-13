package handler

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/stock"
	pagination "market/backend/internal/infrastructure/delivery/webapp/dto/pagination"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/stock"
)

var handlerEntryMovement = domain.StockMovement{
	ID: "m1", ItemID: "i1", Type: domain.MovementEntry, Quantity: 12_000, BalanceAfter: 16_000,
	Reference: ptr("Remito 12"), SupplierID: ptr("s1"), UnitCost: ptr(domain.Money(320_000)),
	ActorID: "u1", CreatedAt: handlerTime,
}

func assertEntryMovement(t *testing.T, got dto.StockMovementResponse) {
	t.Helper()
	if got.ID != "m1" || got.ItemID != "i1" || got.Type != "entry" || got.Quantity != 12_000 || got.BalanceAfter != 16_000 ||
		got.Reference == nil || *got.Reference != "Remito 12" || got.SupplierID == nil || *got.SupplierID != "s1" ||
		got.UnitCost == nil || *got.UnitCost != 320_000 || got.Reason != nil || got.ActorID != "u1" || !got.CreatedAt.Equal(handlerTime) {
		t.Fatalf("movement = %+v", got)
	}
}

func TestListStockLevelsHandler(t *testing.T) {
	var received stock.ListStockLevelsInput
	yerba := domain.ItemStock{
		Item:  domain.Item{ID: "i1", ItemFields: domain.ItemFields{Code: "YER-1", Name: "Yerba", CategoryID: "c1", Unit: domain.UnitPiece, MinStock: 6_000}},
		Level: domain.StockLevel{ItemID: "i1", Quantity: 4_000, UpdatedAt: handlerTime},
	}
	queso := domain.ItemStock{
		Item:  domain.Item{ID: "i2", ItemFields: domain.ItemFields{Code: "QUE-1", Name: "Queso", CategoryID: "c2", Unit: domain.UnitKilogram}},
		Level: domain.StockLevel{ItemID: "i2"},
	}
	listStockLevels := useCaseFunc[stock.ListStockLevelsInput, domain.Page[domain.ItemStock]](
		func(_ context.Context, input stock.ListStockLevelsInput) (domain.Page[domain.ItemStock], error) {
			received = input
			return domain.Page[domain.ItemStock]{Items: []domain.ItemStock{yerba, queso}, Total: 2, Page: 1, Limit: 25}, nil
		})

	recorder := serve(NewListStockLevelsHandler(listStockLevels), categoryRequest(http.MethodGet, "/api/v1/stock?q=yer&categoryId=c1&itemId=i1&itemId=i2&lowStock=true&page=1&limit=25", nil))

	assertStatus(t, recorder, http.StatusOK)
	got := decodeBody[pagination.PageResponse[dto.StockLevelResponse]](t, recorder)
	if got.Total != 2 || len(got.Items) != 2 {
		t.Fatalf("body = %+v", got)
	}
	first, second := got.Items[0], got.Items[1]
	if first.ItemID != "i1" || first.Code != "YER-1" || first.Name != "Yerba" || first.CategoryID != "c1" || first.Unit != "unit" ||
		first.MinStock != 6_000 || first.Quantity != 4_000 || !first.LowStock || first.UpdatedAt == nil || !first.UpdatedAt.Equal(handlerTime) {
		t.Fatalf("first level = %+v", first)
	}
	if second.Quantity != 0 || second.LowStock || second.UpdatedAt != nil || second.Unit != "kg" {
		t.Fatalf("level without movements = %+v", second)
	}
	want := stock.ListStockLevelsInput{Actor: handlerActor, Search: "yer", CategoryID: "c1", ItemIDs: []string{"i1", "i2"}, LowStock: true, Page: 1, Limit: 25}
	if !reflect.DeepEqual(received, want) {
		t.Fatalf("input = %+v, want %+v", received, want)
	}

	serve(NewListStockLevelsHandler(listStockLevels), categoryRequest(http.MethodGet, "/api/v1/stock?lowStock=yes", nil))
	if received.LowStock {
		t.Fatal(`lowStock=yes filtered by low stock, want only "true" to`)
	}
	if received.ItemIDs != nil {
		t.Fatalf("ItemIDs without itemId = %q, want none", received.ItemIDs)
	}
	assertStatus(t, serve(NewListStockLevelsHandler(listStockLevels), categoryRequest(http.MethodGet, "/api/v1/stock?limit=x", nil)), http.StatusBadRequest)
}

func TestListStockLevelsHandlerErrors(t *testing.T) {
	forbidden := useCaseFunc[stock.ListStockLevelsInput, domain.Page[domain.ItemStock]](
		func(context.Context, stock.ListStockLevelsInput) (domain.Page[domain.ItemStock], error) {
			return domain.Page[domain.ItemStock]{}, domain.ErrForbidden
		})

	assertStatus(t, serve(NewListStockLevelsHandler(forbidden), categoryRequest(http.MethodGet, "/api/v1/stock", nil)), http.StatusForbidden)
}

func TestListStockMovementsHandler(t *testing.T) {
	var received stock.ListStockMovementsInput
	listStockMovements := useCaseFunc[stock.ListStockMovementsInput, domain.Page[domain.StockMovement]](
		func(_ context.Context, input stock.ListStockMovementsInput) (domain.Page[domain.StockMovement], error) {
			received = input
			if input.Type == "transfer" {
				return domain.Page[domain.StockMovement]{}, domain.ErrStockMovementTypeInvalid
			}
			return domain.Page[domain.StockMovement]{Items: []domain.StockMovement{handlerEntryMovement}, Total: 1, Page: 2, Limit: 10}, nil
		})

	recorder := serve(NewListStockMovementsHandler(listStockMovements), categoryRequest(http.MethodGet, "/api/v1/stock/movements?itemId=i1&type=entry&page=2&limit=10", nil))

	assertStatus(t, recorder, http.StatusOK)
	got := decodeBody[pagination.PageResponse[dto.StockMovementResponse]](t, recorder)
	if got.Total != 1 || got.Page != 2 || len(got.Items) != 1 {
		t.Fatalf("body = %+v", got)
	}
	assertEntryMovement(t, got.Items[0])
	want := stock.ListStockMovementsInput{Actor: handlerActor, ItemID: "i1", Type: domain.MovementEntry, Page: 2, Limit: 10}
	if received != want {
		t.Fatalf("input = %+v, want %+v", received, want)
	}

	assertStatus(t, serve(NewListStockMovementsHandler(listStockMovements), categoryRequest(http.MethodGet, "/api/v1/stock/movements?type=transfer", nil)), http.StatusBadRequest)
	assertStatus(t, serve(NewListStockMovementsHandler(listStockMovements), categoryRequest(http.MethodGet, "/api/v1/stock/movements?page=x", nil)), http.StatusBadRequest)
}

func TestCreateStockEntryHandler(t *testing.T) {
	var received stock.RegisterStockEntryInput
	registerStockEntry := useCaseFunc[stock.RegisterStockEntryInput, []domain.StockMovement](
		func(_ context.Context, input stock.RegisterStockEntryInput) ([]domain.StockMovement, error) {
			received = input
			if len(input.Entry.Lines) == 0 {
				return nil, domain.ErrStockEntryLinesRequired
			}
			return []domain.StockMovement{handlerEntryMovement}, nil
		})
	body := `{"supplierId":"s1","reference":"Remito 12","lines":[{"itemId":"i1","quantity":12000,"unitCost":320000},{"itemId":"i2","quantity":2500,"unitCost":null}]}`

	recorder := serve(NewCreateStockEntryHandler(registerStockEntry), categoryRequest(http.MethodPost, "/api/v1/stock/entries", strings.NewReader(body)))

	assertStatus(t, recorder, http.StatusCreated)
	got := decodeBody[dto.StockEntryResponse](t, recorder)
	if len(got.Movements) != 1 {
		t.Fatalf("body = %+v", got)
	}
	assertEntryMovement(t, got.Movements[0])
	entry := received.Entry
	if received.Actor != handlerActor || *entry.SupplierID != "s1" || *entry.Reference != "Remito 12" || len(entry.Lines) != 2 ||
		entry.Lines[0] != (domain.StockEntryLine{ItemID: "i1", Quantity: 12_000, UnitCost: entry.Lines[0].UnitCost}) || *entry.Lines[0].UnitCost != 320_000 ||
		entry.Lines[1].ItemID != "i2" || entry.Lines[1].Quantity != 2_500 || entry.Lines[1].UnitCost != nil {
		t.Fatalf("input = %+v", received)
	}

	assertStatus(t, serve(NewCreateStockEntryHandler(registerStockEntry), categoryRequest(http.MethodPost, "/api/v1/stock/entries", strings.NewReader(`{"lines":[]}`))), http.StatusBadRequest)
	assertStatus(t, serve(NewCreateStockEntryHandler(registerStockEntry), categoryRequest(http.MethodPost, "/api/v1/stock/entries", strings.NewReader(`{`))), http.StatusBadRequest)
}

func TestCreateStockAdjustmentHandler(t *testing.T) {
	var received stock.RegisterStockAdjustmentInput
	adjusted := domain.StockMovement{
		ID: "m2", ItemID: "i1", Type: domain.MovementAdjustment, Quantity: -2_000, BalanceAfter: 2_000,
		Reason: ptr("Rotura"), ActorID: "u1", CreatedAt: handlerTime,
	}
	registerStockAdjustment := useCaseFunc[stock.RegisterStockAdjustmentInput, domain.StockMovement](
		func(_ context.Context, input stock.RegisterStockAdjustmentInput) (domain.StockMovement, error) {
			received = input
			if input.Adjustment.Quantity < -4_000 {
				return domain.StockMovement{}, domain.InsufficientStock("Yerba")
			}
			return adjusted, nil
		})

	recorder := serve(NewCreateStockAdjustmentHandler(registerStockAdjustment), categoryRequest(http.MethodPost, "/api/v1/stock/adjustments",
		strings.NewReader(`{"itemId":"i1","quantity":-2000,"reason":"Rotura"}`)))

	assertStatus(t, recorder, http.StatusCreated)
	got := decodeBody[dto.StockMovementResponse](t, recorder)
	if got.ID != "m2" || got.Type != "adjustment" || got.Quantity != -2_000 || got.BalanceAfter != 2_000 ||
		got.Reason == nil || *got.Reason != "Rotura" || got.SupplierID != nil || got.UnitCost != nil {
		t.Fatalf("body = %+v", got)
	}
	want := stock.RegisterStockAdjustmentInput{Actor: handlerActor, Adjustment: domain.StockAdjustment{ItemID: "i1", Quantity: -2_000, Reason: "Rotura"}}
	if received != want {
		t.Fatalf("input = %+v, want %+v", received, want)
	}

	short := serve(NewCreateStockAdjustmentHandler(registerStockAdjustment), categoryRequest(http.MethodPost, "/api/v1/stock/adjustments",
		strings.NewReader(`{"itemId":"i1","quantity":-5000,"reason":"Rotura"}`)))
	assertStatus(t, short, http.StatusConflict)
	if !strings.Contains(short.Body.String(), `"code":"insufficient_stock"`) || !strings.Contains(short.Body.String(), "Yerba") {
		t.Fatalf("body = %s", short.Body.String())
	}
	assertStatus(t, serve(NewCreateStockAdjustmentHandler(registerStockAdjustment), categoryRequest(http.MethodPost, "/api/v1/stock/adjustments", strings.NewReader(`[`))), http.StatusBadRequest)
}
