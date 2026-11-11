package repositorytest

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

func newItem(t *testing.T, categoryID, code, name string) domain.Item {
	t.Helper()
	item, err := domain.NewItem(domain.ItemFields{
		Code:       code,
		Name:       name,
		CategoryID: categoryID,
		Unit:       domain.UnitPiece,
		SalePrice:  100_000,
		TaxRate:    2100,
	}, contractTime)
	if err != nil {
		t.Fatalf("domain.NewItem(%q) error = %v", code, err)
	}
	return item
}

func mustCreateItem(t *testing.T, repository gateway.ItemRepository, item domain.Item) domain.Item {
	t.Helper()
	created, err := repository.Create(context.Background(), item)
	if err != nil {
		t.Fatalf("Create(%q) error = %v", item.Code, err)
	}
	return created
}

func itemNames(items []domain.Item) []string {
	result := make([]string, len(items))
	for i, item := range items {
		result[i] = item.Name
	}
	return result
}

// ItemRepository runs the gateway.ItemRepository contract.
func ItemRepository(t *testing.T, newRepositories NewRepositories) {
	ctx := context.Background()

	setup := func(t *testing.T) (gateway.ItemRepository, domain.Category, domain.Category) {
		t.Helper()
		repositories := newRepositories(t)
		almacen := mustCreate(t, repositories.Categories, newCategory(t, "Almacén", nil))
		bebidas := mustCreate(t, repositories.Categories, newCategory(t, "Bebidas", nil))
		return repositories.Items, almacen, bebidas
	}

	t.Run("create assigns an id and get returns every field", func(t *testing.T) {
		items, almacen, _ := setup(t)
		item := newItem(t, almacen.ID, "YER-1", "Yerba mate 1 kg")
		item.Barcode = text("7790387000123")
		item.Description = text("Con palo")
		cost := domain.Money(60_000)
		item.Cost = &cost
		item.Unit = domain.UnitKilogram
		item.TracksStock = true
		item.MinStock = 2500

		created := mustCreateItem(t, items, item)
		if created.ID == "" {
			t.Fatal("Create did not assign an ID")
		}

		got, err := items.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.Code != "YER-1" || got.Name != "Yerba mate 1 kg" || got.CategoryID != almacen.ID ||
			got.Barcode == nil || *got.Barcode != "7790387000123" ||
			got.Description == nil || *got.Description != "Con palo" ||
			got.Cost == nil || *got.Cost != 60_000 ||
			got.Unit != domain.UnitKilogram || got.SalePrice != 100_000 || got.TaxRate != 2100 ||
			!got.TracksStock || got.MinStock != 2500 || !got.CreatedAt.Equal(contractTime) {
			t.Fatalf("Get() = %+v", got)
		}
	})

	t.Run("code is unique ignoring case and barcode is unique", func(t *testing.T) {
		items, almacen, _ := setup(t)
		first := newItem(t, almacen.ID, "YER-1", "Yerba")
		first.Barcode = text("7790387000123")
		mustCreateItem(t, items, first)

		if _, err := items.Create(ctx, newItem(t, almacen.ID, "yer-1", "Otra yerba")); !errors.Is(err, domain.ErrItemCodeInUse) {
			t.Fatalf("Create(duplicate code) error = %v, want ErrItemCodeInUse", err)
		}

		sameBarcode := newItem(t, almacen.ID, "YER-2", "Otra yerba")
		sameBarcode.Barcode = text("7790387000123")
		if _, err := items.Create(ctx, sameBarcode); !errors.Is(err, domain.ErrItemBarcodeInUse) {
			t.Fatalf("Create(duplicate barcode) error = %v, want ErrItemBarcodeInUse", err)
		}

		mustCreateItem(t, items, newItem(t, almacen.ID, "YER-3", "Sin código de barras"))
		mustCreateItem(t, items, newItem(t, almacen.ID, "YER-4", "Tampoco tiene"))
	})

	t.Run("the category must exist", func(t *testing.T) {
		items, almacen, _ := setup(t)

		if _, err := items.Create(ctx, newItem(t, "00000000-0000-7000-8000-000000000000", "X-1", "Huérfano")); !errors.Is(err, domain.ErrItemCategoryNotFound) {
			t.Fatalf("Create(unknown category) error = %v, want ErrItemCategoryNotFound", err)
		}

		created := mustCreateItem(t, items, newItem(t, almacen.ID, "X-2", "Con categoría"))
		created.CategoryID = "00000000-0000-7000-8000-000000000000"
		if _, err := items.Update(ctx, created); !errors.Is(err, domain.ErrItemCategoryNotFound) {
			t.Fatalf("Update(unknown category) error = %v, want ErrItemCategoryNotFound", err)
		}
	})

	t.Run("update replaces the item and keeps the constraints", func(t *testing.T) {
		items, almacen, bebidas := setup(t)
		yerba := mustCreateItem(t, items, newItem(t, almacen.ID, "YER-1", "Yerba"))
		mustCreateItem(t, items, newItem(t, almacen.ID, "COC-1", "Coca"))

		later := contractTime.Add(time.Hour)
		fields := yerba.ItemFields
		fields.Name = "Yerba suave"
		fields.CategoryID = bebidas.ID
		changed, err := yerba.Replace(fields, later)
		if err != nil {
			t.Fatalf("Replace() error = %v", err)
		}
		if _, err := items.Update(ctx, changed); err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		got, _ := items.Get(ctx, yerba.ID)
		if got.Name != "Yerba suave" || got.CategoryID != bebidas.ID || !got.UpdatedAt.Equal(later) {
			t.Fatalf("Get() after update = %+v", got)
		}

		got.Code = "coc-1"
		if _, err := items.Update(ctx, got); !errors.Is(err, domain.ErrItemCodeInUse) {
			t.Fatalf("Update(into a taken code) error = %v, want ErrItemCodeInUse", err)
		}
	})

	t.Run("delete removes the item and unknown ids are not found", func(t *testing.T) {
		items, almacen, _ := setup(t)
		yerba := mustCreateItem(t, items, newItem(t, almacen.ID, "YER-1", "Yerba"))

		if err := items.Delete(ctx, yerba.ID); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
		if _, err := items.Get(ctx, yerba.ID); !errors.Is(err, domain.ErrItemNotFound) {
			t.Fatalf("Get(deleted) error = %v, want ErrItemNotFound", err)
		}
		if err := items.Delete(ctx, yerba.ID); !errors.Is(err, domain.ErrItemNotFound) {
			t.Fatalf("Delete(deleted) error = %v, want ErrItemNotFound", err)
		}
		if _, err := items.Update(ctx, yerba); !errors.Is(err, domain.ErrItemNotFound) {
			t.Fatalf("Update(deleted) error = %v, want ErrItemNotFound", err)
		}
		mustCreateItem(t, items, newItem(t, almacen.ID, "YER-1", "Yerba otra vez"))
	})

	t.Run("list searches by name, code and barcode, filters by category, sorted by name", func(t *testing.T) {
		items, almacen, bebidas := setup(t)
		yerba := newItem(t, almacen.ID, "YER-1", "Yerba mate")
		yerba.Barcode = text("7790387000123")
		mustCreateItem(t, items, yerba)
		mustCreateItem(t, items, newItem(t, almacen.ID, "ACE-1", "aceite de girasol"))
		mustCreateItem(t, items, newItem(t, bebidas.ID, "COC-1", "Coca-Cola 500 ml"))
		page := domain.PageRequest{Page: 1, Limit: 10}

		tests := []struct {
			name   string
			filter gateway.ItemFilter
			want   []string
		}{
			{name: "all", filter: gateway.ItemFilter{Page: page}, want: []string{"aceite de girasol", "Coca-Cola 500 ml", "Yerba mate"}},
			{name: "by category", filter: gateway.ItemFilter{CategoryID: almacen.ID, Page: page}, want: []string{"aceite de girasol", "Yerba mate"}},
			{name: "by name without accents", filter: gateway.ItemFilter{Search: "GIRASOL", Page: page}, want: []string{"aceite de girasol"}},
			{name: "by code", filter: gateway.ItemFilter{Search: "coc-1", Page: page}, want: []string{"Coca-Cola 500 ml"}},
			{name: "by barcode", filter: gateway.ItemFilter{Search: "7790387", Page: page}, want: []string{"Yerba mate"}},
			{name: "search within category", filter: gateway.ItemFilter{Search: "coca", CategoryID: almacen.ID, Page: page}, want: []string{}},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				got, err := items.List(ctx, test.filter)
				if err != nil {
					t.Fatalf("List() error = %v", err)
				}
				if fmt.Sprint(itemNames(got.Items)) != fmt.Sprint(test.want) || got.Total != len(test.want) || got.Items == nil {
					t.Fatalf("List() = %v (total %d), want %v", itemNames(got.Items), got.Total, test.want)
				}
			})
		}
	})

	t.Run("list pages through the results", func(t *testing.T) {
		items, almacen, _ := setup(t)
		for _, name := range []string{"E", "A", "D", "C", "B"} {
			mustCreateItem(t, items, newItem(t, almacen.ID, "CODE-"+name, name))
		}

		second, err := items.List(ctx, gateway.ItemFilter{Page: domain.PageRequest{Page: 2, Limit: 2}})
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if fmt.Sprint(itemNames(second.Items)) != "[C D]" || second.Total != 5 || second.Page != 2 || second.Limit != 2 {
			t.Fatalf("page 2 = %v %+v", itemNames(second.Items), second)
		}
	})

	t.Run("callers cannot mutate stored items", func(t *testing.T) {
		items, almacen, _ := setup(t)
		item := newItem(t, almacen.ID, "YER-1", "Yerba")
		item.Barcode = text("123")
		item.Description = text("Original")
		cost := domain.Money(500)
		item.Cost = &cost
		created := mustCreateItem(t, items, item)

		*created.Barcode, *created.Description, *created.Cost = "999", "changed", 1

		again, _ := items.Get(ctx, created.ID)
		if *again.Barcode != "123" || *again.Description != "Original" || *again.Cost != 500 {
			t.Fatalf("stored item = %+v", again)
		}
	})
}
