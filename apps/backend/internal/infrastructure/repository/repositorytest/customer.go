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

func newCustomer(t *testing.T, name string, documentType domain.DocumentType, number string) domain.Customer {
	t.Helper()
	customer, err := domain.NewCustomer(domain.CustomerFields{
		Name:           name,
		DocumentType:   documentType,
		DocumentNumber: number,
		TaxCondition:   domain.TaxFinalConsumer,
	}, contractTime)
	if err != nil {
		t.Fatalf("domain.NewCustomer(%q) error = %v", name, err)
	}
	return customer
}

func mustCreateCustomer(t *testing.T, repository gateway.CustomerRepository, customer domain.Customer) domain.Customer {
	t.Helper()
	created, err := repository.Create(context.Background(), customer)
	if err != nil {
		t.Fatalf("Create(%q) error = %v", customer.Name, err)
	}
	return created
}

func customerNames(customers []domain.Customer) []string {
	result := make([]string, len(customers))
	for i, customer := range customers {
		result[i] = customer.Name
	}
	return result
}

// CustomerRepository runs the gateway.CustomerRepository contract.
func CustomerRepository(t *testing.T, newRepositories NewRepositories) {
	ctx := context.Background()
	newRepository := func(t *testing.T) gateway.CustomerRepository {
		return newRepositories(t).Customers
	}

	t.Run("create assigns an id and get returns every field", func(t *testing.T) {
		repository := newRepository(t)
		customer := newCustomer(t, "La Esquina SRL", domain.DocumentCUIT, "30-71234567-8")
		customer.TaxCondition = domain.TaxRegistered
		customer.Email = text("compras@laesquina.com.ar")
		customer.Phone = text("351 555-1234")
		customer.Address = text("San Martín 123")
		customer.Notes = text("Paga a 30 días")

		created := mustCreateCustomer(t, repository, customer)
		if created.ID == "" {
			t.Fatal("Create did not assign an ID")
		}

		got, err := repository.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.Name != "La Esquina SRL" || got.DocumentType != domain.DocumentCUIT || got.DocumentNumber != "30712345678" ||
			got.TaxCondition != domain.TaxRegistered || *got.Email != "compras@laesquina.com.ar" ||
			*got.Phone != "351 555-1234" || *got.Address != "San Martín 123" || *got.Notes != "Paga a 30 días" ||
			!got.CreatedAt.Equal(contractTime) {
			t.Fatalf("Get() = %+v", got)
		}
	})

	t.Run("a document is unique per type, and customers without one never collide", func(t *testing.T) {
		repository := newRepository(t)
		mustCreateCustomer(t, repository, newCustomer(t, "Ana", domain.DocumentDNI, "12345678"))

		if _, err := repository.Create(ctx, newCustomer(t, "Otra Ana", domain.DocumentDNI, "12.345.678")); !errors.Is(err, domain.ErrCustomerDocumentInUse) {
			t.Fatalf("Create(duplicate document) error = %v, want ErrCustomerDocumentInUse", err)
		}

		mustCreateCustomer(t, repository, newCustomer(t, "Juan", domain.DocumentNone, ""))
		mustCreateCustomer(t, repository, newCustomer(t, "Pedro", domain.DocumentNone, ""))
	})

	t.Run("update replaces the customer and keeps the document check", func(t *testing.T) {
		repository := newRepository(t)
		ana := mustCreateCustomer(t, repository, newCustomer(t, "Ana", domain.DocumentDNI, "12345678"))
		mustCreateCustomer(t, repository, newCustomer(t, "Beto", domain.DocumentDNI, "87654321"))

		later := contractTime.Add(time.Hour)
		fields := ana.CustomerFields
		fields.Name = "Ana María"
		fields.Phone = text("351 111-2222")
		changed, err := ana.Replace(fields, later)
		if err != nil {
			t.Fatalf("Replace() error = %v", err)
		}
		if _, err := repository.Update(ctx, changed); err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		got, _ := repository.Get(ctx, ana.ID)
		if got.Name != "Ana María" || got.Phone == nil || *got.Phone != "351 111-2222" || !got.UpdatedAt.Equal(later) {
			t.Fatalf("Get() after update = %+v", got)
		}

		got.DocumentNumber = "87654321"
		if _, err := repository.Update(ctx, got); !errors.Is(err, domain.ErrCustomerDocumentInUse) {
			t.Fatalf("Update(into a taken document) error = %v, want ErrCustomerDocumentInUse", err)
		}
	})

	t.Run("delete removes the customer and unknown ids are not found", func(t *testing.T) {
		repository := newRepository(t)
		ana := mustCreateCustomer(t, repository, newCustomer(t, "Ana", domain.DocumentDNI, "12345678"))

		if err := repository.Delete(ctx, ana.ID); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
		if _, err := repository.Get(ctx, ana.ID); !errors.Is(err, domain.ErrCustomerNotFound) {
			t.Fatalf("Get(deleted) error = %v, want ErrCustomerNotFound", err)
		}
		if err := repository.Delete(ctx, ana.ID); !errors.Is(err, domain.ErrCustomerNotFound) {
			t.Fatalf("Delete(deleted) error = %v, want ErrCustomerNotFound", err)
		}
		if _, err := repository.Update(ctx, ana); !errors.Is(err, domain.ErrCustomerNotFound) {
			t.Fatalf("Update(deleted) error = %v, want ErrCustomerNotFound", err)
		}
		mustCreateCustomer(t, repository, newCustomer(t, "Ana otra vez", domain.DocumentDNI, "12345678"))
	})

	t.Run("list searches by name, document, email and phone, sorted by name", func(t *testing.T) {
		repository := newRepository(t)
		esquina := newCustomer(t, "La Esquina SRL", domain.DocumentCUIT, "30712345678")
		esquina.TaxCondition = domain.TaxRegistered
		esquina.Email = text("compras@laesquina.com.ar")
		mustCreateCustomer(t, repository, esquina)
		ana := newCustomer(t, "ana Pérez", domain.DocumentDNI, "40111222")
		ana.Phone = text("351 555-0000")
		mustCreateCustomer(t, repository, ana)
		mustCreateCustomer(t, repository, newCustomer(t, "Beto", domain.DocumentNone, ""))
		page := domain.PageRequest{Page: 1, Limit: 10}

		tests := []struct {
			search string
			want   []string
		}{
			{search: "", want: []string{"ana Pérez", "Beto", "La Esquina SRL"}},
			{search: "PEREZ", want: []string{"ana Pérez"}},
			{search: "30-71234567", want: []string{"La Esquina SRL"}},
			{search: "40.111", want: []string{"ana Pérez"}},
			{search: "laesquina", want: []string{"La Esquina SRL"}},
			{search: "555-0000", want: []string{"ana Pérez"}},
			{search: "nadie", want: []string{}},
		}
		for _, test := range tests {
			t.Run(test.search, func(t *testing.T) {
				got, err := repository.List(ctx, gateway.CustomerFilter{Search: test.search, Page: page})
				if err != nil {
					t.Fatalf("List() error = %v", err)
				}
				if fmt.Sprint(customerNames(got.Items)) != fmt.Sprint(test.want) || got.Total != len(test.want) || got.Items == nil {
					t.Fatalf("List(%q) = %v (total %d), want %v", test.search, customerNames(got.Items), got.Total, test.want)
				}
			})
		}
	})

	t.Run("list pages through the results", func(t *testing.T) {
		repository := newRepository(t)
		for _, name := range []string{"E", "A", "D", "C", "B"} {
			mustCreateCustomer(t, repository, newCustomer(t, name, domain.DocumentNone, ""))
		}

		second, err := repository.List(ctx, gateway.CustomerFilter{Page: domain.PageRequest{Page: 2, Limit: 2}})
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if fmt.Sprint(customerNames(second.Items)) != "[C D]" || second.Total != 5 {
			t.Fatalf("page 2 = %v (total %d)", customerNames(second.Items), second.Total)
		}
	})

	t.Run("callers cannot mutate stored customers", func(t *testing.T) {
		repository := newRepository(t)
		customer := newCustomer(t, "Ana", domain.DocumentNone, "")
		customer.Email, customer.Phone = text("ana@x.com"), text("123")
		customer.Address, customer.Notes = text("Calle 1"), text("Original")
		created := mustCreateCustomer(t, repository, customer)

		*created.Email, *created.Phone, *created.Address, *created.Notes = "x", "x", "x", "x"

		again, _ := repository.Get(ctx, created.ID)
		if *again.Email != "ana@x.com" || *again.Phone != "123" || *again.Address != "Calle 1" || *again.Notes != "Original" {
			t.Fatalf("stored customer = %+v", again)
		}
	})

	t.Run("concurrent creates with the same document keep only one", func(t *testing.T) {
		repository := newRepository(t)
		ana := newCustomer(t, "Ana", domain.DocumentDNI, "12345678")
		var wg sync.WaitGroup
		var mu sync.Mutex
		succeeded := 0

		for range 20 {
			wg.Go(func() {
				if _, err := repository.Create(ctx, ana); err == nil {
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
