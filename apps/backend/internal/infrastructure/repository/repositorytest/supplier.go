package repositorytest

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

func newSupplier(t *testing.T, businessName string, cuit *string) domain.Supplier {
	t.Helper()
	supplier, err := domain.NewSupplier(domain.SupplierFields{BusinessName: businessName, CUIT: cuit}, contractTime)
	if err != nil {
		t.Fatalf("domain.NewSupplier(%q) error = %v", businessName, err)
	}
	return supplier
}

func mustCreateSupplier(t *testing.T, repository gateway.SupplierRepository, supplier domain.Supplier) domain.Supplier {
	t.Helper()
	created, err := repository.Create(context.Background(), supplier)
	if err != nil {
		t.Fatalf("Create(%q) error = %v", supplier.BusinessName, err)
	}
	return created
}

func supplierNames(suppliers []domain.Supplier) []string {
	result := make([]string, len(suppliers))
	for i, supplier := range suppliers {
		result[i] = supplier.BusinessName
	}
	return result
}

// SupplierRepository runs the gateway.SupplierRepository contract.
func SupplierRepository(t *testing.T, newRepositories NewRepositories) {
	ctx := context.Background()
	newRepository := func(t *testing.T) gateway.SupplierRepository {
		return newRepositories(t).Suppliers
	}

	t.Run("create assigns an id and get returns every field", func(t *testing.T) {
		repository := newRepository(t)
		supplier := newSupplier(t, "Distribuidora del Centro SA", text("30-70123456-1"))
		supplier.ContactName = text("Laura Díaz")
		supplier.Phone = text("351 555-2000")
		supplier.Email = text("pedidos@delcentro.com.ar")
		supplier.Address = text("Av. Colón 1500")
		supplier.Notes = text("Entrega martes y viernes")

		created := mustCreateSupplier(t, repository, supplier)
		if created.ID == "" {
			t.Fatal("Create did not assign an ID")
		}

		got, err := repository.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.BusinessName != "Distribuidora del Centro SA" || *got.CUIT != "30701234561" || *got.ContactName != "Laura Díaz" ||
			*got.Phone != "351 555-2000" || *got.Email != "pedidos@delcentro.com.ar" || *got.Address != "Av. Colón 1500" ||
			*got.Notes != "Entrega martes y viernes" || !got.CreatedAt.Equal(contractTime) {
			t.Fatalf("Get() = %+v", got)
		}
	})

	t.Run("a CUIT is unique, and suppliers without one never collide", func(t *testing.T) {
		repository := newRepository(t)
		mustCreateSupplier(t, repository, newSupplier(t, "Distribuidora del Centro", text("30701234561")))

		if _, err := repository.Create(ctx, newSupplier(t, "Otra distribuidora", text("30-70123456-1"))); !errors.Is(err, domain.ErrSupplierCUITInUse) {
			t.Fatalf("Create(duplicate CUIT) error = %v, want ErrSupplierCUITInUse", err)
		}

		mustCreateSupplier(t, repository, newSupplier(t, "Verdulería", nil))
		mustCreateSupplier(t, repository, newSupplier(t, "Panadería", nil))
	})

	t.Run("update replaces the supplier and keeps the CUIT check", func(t *testing.T) {
		repository := newRepository(t)
		centro := mustCreateSupplier(t, repository, newSupplier(t, "Distribuidora del Centro", text("30701234561")))
		mustCreateSupplier(t, repository, newSupplier(t, "Lácteos del Sur", text("30709876543")))

		later := contractTime.Add(time.Hour)
		fields := centro.SupplierFields
		fields.BusinessName = "Distribuidora del Centro SA"
		fields.Phone = text("351 111-2222")
		changed, err := centro.Replace(fields, later)
		if err != nil {
			t.Fatalf("Replace() error = %v", err)
		}
		if _, err := repository.Update(ctx, changed); err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		got, _ := repository.Get(ctx, centro.ID)
		if got.BusinessName != "Distribuidora del Centro SA" || got.Phone == nil || *got.Phone != "351 111-2222" || !got.UpdatedAt.Equal(later) {
			t.Fatalf("Get() after update = %+v", got)
		}

		got.CUIT = text("30709876543")
		if _, err := repository.Update(ctx, got); !errors.Is(err, domain.ErrSupplierCUITInUse) {
			t.Fatalf("Update(into a taken CUIT) error = %v, want ErrSupplierCUITInUse", err)
		}
	})

	t.Run("delete removes the supplier and unknown ids are not found", func(t *testing.T) {
		repository := newRepository(t)
		centro := mustCreateSupplier(t, repository, newSupplier(t, "Distribuidora del Centro", text("30701234561")))

		if err := repository.Delete(ctx, centro.ID); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
		if _, err := repository.Get(ctx, centro.ID); !errors.Is(err, domain.ErrSupplierNotFound) {
			t.Fatalf("Get(deleted) error = %v, want ErrSupplierNotFound", err)
		}
		if err := repository.Delete(ctx, centro.ID); !errors.Is(err, domain.ErrSupplierNotFound) {
			t.Fatalf("Delete(deleted) error = %v, want ErrSupplierNotFound", err)
		}
		if _, err := repository.Update(ctx, centro); !errors.Is(err, domain.ErrSupplierNotFound) {
			t.Fatalf("Update(deleted) error = %v, want ErrSupplierNotFound", err)
		}
		mustCreateSupplier(t, repository, newSupplier(t, "Distribuidora del Centro", text("30701234561")))
	})

	t.Run("a supplier listed by items can't be deleted", func(t *testing.T) {
		repositories := newRepositories(t)
		almacen := mustCreate(t, repositories.Categories, newCategory(t, "Almacén", nil))
		centro := mustCreateSupplier(t, repositories.Suppliers, newSupplier(t, "Distribuidora del Centro", nil))
		item := newItem(t, almacen.ID, "YER-1", "Yerba")
		item.SupplierIDs = []string{centro.ID}
		created := mustCreateItem(t, repositories.Items, item)

		if err := repositories.Suppliers.Delete(ctx, centro.ID); !errors.Is(err, domain.ErrSupplierInUse) {
			t.Fatalf("Delete(supplier with items) error = %v, want ErrSupplierInUse", err)
		}

		created.SupplierIDs = nil
		if _, err := repositories.Items.Update(ctx, created); err != nil {
			t.Fatalf("Items.Update() error = %v", err)
		}
		if err := repositories.Suppliers.Delete(ctx, centro.ID); err != nil {
			t.Fatalf("Delete(supplier without items) error = %v", err)
		}
	})

	t.Run("list searches by business name, contact, CUIT, email and phone, sorted by business name", func(t *testing.T) {
		repository := newRepository(t)
		centro := newSupplier(t, "Distribuidora del Centro", text("30701234561"))
		centro.ContactName = text("Laura Díaz")
		centro.Email = text("pedidos@delcentro.com.ar")
		mustCreateSupplier(t, repository, centro)
		lacteos := newSupplier(t, "lácteos del Sur", text("30709876543"))
		lacteos.Phone = text("351 555-0000")
		mustCreateSupplier(t, repository, lacteos)
		mustCreateSupplier(t, repository, newSupplier(t, "Verdulería", nil))
		page := domain.PageRequest{Page: 1, Limit: 10}

		tests := []struct {
			search string
			want   []string
		}{
			{search: "", want: []string{"Distribuidora del Centro", "lácteos del Sur", "Verdulería"}},
			{search: "LACTEOS", want: []string{"lácteos del Sur"}},
			{search: "diaz", want: []string{"Distribuidora del Centro"}},
			{search: "30-70123", want: []string{"Distribuidora del Centro"}},
			{search: "delcentro", want: []string{"Distribuidora del Centro"}},
			{search: "555-0000", want: []string{"lácteos del Sur"}},
			{search: "del", want: []string{"Distribuidora del Centro", "lácteos del Sur"}},
			{search: "nadie", want: []string{}},
		}
		for _, test := range tests {
			t.Run(test.search, func(t *testing.T) {
				got, err := repository.List(ctx, gateway.SupplierFilter{Search: test.search, Page: page})
				if err != nil {
					t.Fatalf("List() error = %v", err)
				}
				if fmt.Sprint(supplierNames(got.Items)) != fmt.Sprint(test.want) || got.Total != len(test.want) || got.Items == nil {
					t.Fatalf("List(%q) = %v (total %d), want %v", test.search, supplierNames(got.Items), got.Total, test.want)
				}
			})
		}
	})

	t.Run("list pages through the results", func(t *testing.T) {
		repository := newRepository(t)
		for _, name := range []string{"E", "A", "D", "C", "B"} {
			mustCreateSupplier(t, repository, newSupplier(t, name, nil))
		}

		second, err := repository.List(ctx, gateway.SupplierFilter{Page: domain.PageRequest{Page: 2, Limit: 2}})
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if fmt.Sprint(supplierNames(second.Items)) != "[C D]" || second.Total != 5 {
			t.Fatalf("page 2 = %v (total %d)", supplierNames(second.Items), second.Total)
		}
	})

	t.Run("callers cannot mutate stored suppliers", func(t *testing.T) {
		repository := newRepository(t)
		supplier := newSupplier(t, "Centro", text("30701234561"))
		supplier.ContactName, supplier.Phone, supplier.Email = text("Laura"), text("123"), text("a@x.com")
		supplier.Address, supplier.Notes = text("Calle 1"), text("Original")
		created := mustCreateSupplier(t, repository, supplier)

		*created.CUIT, *created.ContactName, *created.Phone = "x", "x", "x"
		*created.Email, *created.Address, *created.Notes = "x", "x", "x"

		again, _ := repository.Get(ctx, created.ID)
		if *again.CUIT != "30701234561" || *again.ContactName != "Laura" || *again.Phone != "123" ||
			*again.Email != "a@x.com" || *again.Address != "Calle 1" || *again.Notes != "Original" {
			t.Fatalf("stored supplier = %+v", again)
		}
	})

	t.Run("concurrent creates with the same CUIT keep only one", func(t *testing.T) {
		repository := newRepository(t)
		centro := newSupplier(t, "Centro", text("30701234561"))
		var wg sync.WaitGroup
		var mu sync.Mutex
		succeeded := 0

		for range 20 {
			wg.Go(func() {
				if _, err := repository.Create(ctx, centro); err == nil {
					mu.Lock()
					succeeded++
					mu.Unlock()
				}
			})
		}
		wg.Wait()

		if succeeded != 1 {
			t.Fatalf("%d concurrent creates succeeded, want 1", succeeded)
		}
	})
}
