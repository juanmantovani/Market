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

func newSeller(t *testing.T, name, code string) domain.Seller {
	t.Helper()
	seller, err := domain.NewSeller(domain.SellerFields{Name: name, Code: code}, contractTime)
	if err != nil {
		t.Fatalf("domain.NewSeller(%q) error = %v", name, err)
	}
	return seller
}

func mustCreateSeller(t *testing.T, repository gateway.SellerRepository, seller domain.Seller) domain.Seller {
	t.Helper()
	created, err := repository.Create(context.Background(), seller)
	if err != nil {
		t.Fatalf("Create(%q) error = %v", seller.Name, err)
	}
	return created
}

func sellerNames(sellers []domain.Seller) []string {
	result := make([]string, len(sellers))
	for i, seller := range sellers {
		result[i] = seller.Name
	}
	return result
}

// SellerRepository runs the gateway.SellerRepository contract.
func SellerRepository(t *testing.T, newRepositories NewRepositories) {
	ctx := context.Background()
	newRepository := func(t *testing.T) gateway.SellerRepository {
		return newRepositories(t).Sellers
	}

	t.Run("create assigns an id and get returns every field", func(t *testing.T) {
		repository := newRepository(t)
		seller := newSeller(t, "Juan Pérez", "jp")
		rate := domain.Percent(250)
		seller.CommissionRate = &rate

		created := mustCreateSeller(t, repository, seller)
		if created.ID == "" {
			t.Fatal("Create did not assign an ID")
		}

		got, err := repository.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.Name != "Juan Pérez" || got.Code != "JP" || got.CommissionRate == nil || *got.CommissionRate != 250 ||
			!got.CreatedAt.Equal(contractTime) {
			t.Fatalf("Get() = %+v", got)
		}
	})

	t.Run("a code is unique", func(t *testing.T) {
		repository := newRepository(t)
		mustCreateSeller(t, repository, newSeller(t, "Juan Pérez", "JP"))

		if _, err := repository.Create(ctx, newSeller(t, "Juana Paz", "jp")); !errors.Is(err, domain.ErrSellerCodeInUse) {
			t.Fatalf("Create(duplicate code) error = %v, want ErrSellerCodeInUse", err)
		}
	})

	t.Run("update replaces the seller and keeps the code check", func(t *testing.T) {
		repository := newRepository(t)
		juan := mustCreateSeller(t, repository, newSeller(t, "Juan", "JP"))
		mustCreateSeller(t, repository, newSeller(t, "Ana", "AN"))

		later := contractTime.Add(time.Hour)
		fields := juan.SellerFields
		fields.Name = "Juan Pablo"
		rate := domain.Percent(300)
		fields.CommissionRate = &rate
		changed, err := juan.Replace(fields, later)
		if err != nil {
			t.Fatalf("Replace() error = %v", err)
		}
		if _, err := repository.Update(ctx, changed); err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		got, _ := repository.Get(ctx, juan.ID)
		if got.Name != "Juan Pablo" || got.CommissionRate == nil || *got.CommissionRate != 300 || !got.UpdatedAt.Equal(later) {
			t.Fatalf("Get() after update = %+v", got)
		}

		got.Code = "AN"
		if _, err := repository.Update(ctx, got); !errors.Is(err, domain.ErrSellerCodeInUse) {
			t.Fatalf("Update(into a taken code) error = %v, want ErrSellerCodeInUse", err)
		}
	})

	t.Run("delete removes the seller and unknown ids are not found", func(t *testing.T) {
		repository := newRepository(t)
		juan := mustCreateSeller(t, repository, newSeller(t, "Juan", "JP"))

		if err := repository.Delete(ctx, juan.ID); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
		if _, err := repository.Get(ctx, juan.ID); !errors.Is(err, domain.ErrSellerNotFound) {
			t.Fatalf("Get(deleted) error = %v, want ErrSellerNotFound", err)
		}
		if err := repository.Delete(ctx, juan.ID); !errors.Is(err, domain.ErrSellerNotFound) {
			t.Fatalf("Delete(deleted) error = %v, want ErrSellerNotFound", err)
		}
		if _, err := repository.Update(ctx, juan); !errors.Is(err, domain.ErrSellerNotFound) {
			t.Fatalf("Update(deleted) error = %v, want ErrSellerNotFound", err)
		}
		mustCreateSeller(t, repository, newSeller(t, "Juan otra vez", "JP"))
	})

	t.Run("list searches by name and code, sorted by name", func(t *testing.T) {
		repository := newRepository(t)
		mustCreateSeller(t, repository, newSeller(t, "Martín Gómez", "MG"))
		mustCreateSeller(t, repository, newSeller(t, "ana Pérez", "AP"))
		mustCreateSeller(t, repository, newSeller(t, "Lucía Paz", "LP2"))
		page := domain.PageRequest{Page: 1, Limit: 10}

		tests := []struct {
			search string
			want   []string
		}{
			{search: "", want: []string{"ana Pérez", "Lucía Paz", "Martín Gómez"}},
			{search: "PEREZ", want: []string{"ana Pérez"}},
			{search: "martin", want: []string{"Martín Gómez"}},
			{search: "lp2", want: []string{"Lucía Paz"}},
			{search: "p", want: []string{"ana Pérez", "Lucía Paz"}},
			{search: "nadie", want: []string{}},
		}
		for _, test := range tests {
			t.Run(test.search, func(t *testing.T) {
				got, err := repository.List(ctx, gateway.SellerFilter{Search: test.search, Page: page})
				if err != nil {
					t.Fatalf("List() error = %v", err)
				}
				if fmt.Sprint(sellerNames(got.Items)) != fmt.Sprint(test.want) || got.Total != len(test.want) || got.Items == nil {
					t.Fatalf("List(%q) = %v (total %d), want %v", test.search, sellerNames(got.Items), got.Total, test.want)
				}
			})
		}
	})

	t.Run("list pages through the results", func(t *testing.T) {
		repository := newRepository(t)
		for _, name := range []string{"E", "A", "D", "C", "B"} {
			mustCreateSeller(t, repository, newSeller(t, name, name))
		}

		second, err := repository.List(ctx, gateway.SellerFilter{Page: domain.PageRequest{Page: 2, Limit: 2}})
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if fmt.Sprint(sellerNames(second.Items)) != "[C D]" || second.Total != 5 {
			t.Fatalf("page 2 = %v (total %d)", sellerNames(second.Items), second.Total)
		}
	})

	t.Run("callers cannot mutate stored sellers", func(t *testing.T) {
		repository := newRepository(t)
		seller := newSeller(t, "Juan", "JP")
		rate := domain.Percent(250)
		seller.CommissionRate = &rate
		created := mustCreateSeller(t, repository, seller)

		rate = 9999
		*created.CommissionRate = 9999

		again, _ := repository.Get(ctx, created.ID)
		if *again.CommissionRate != 250 {
			t.Fatalf("stored seller = %+v", again)
		}
	})

	t.Run("concurrent creates with the same code keep only one", func(t *testing.T) {
		repository := newRepository(t)
		juan := newSeller(t, "Juan", "JP")
		var wg sync.WaitGroup
		var mu sync.Mutex
		succeeded := 0

		for range 20 {
			wg.Go(func() {
				if _, err := repository.Create(ctx, juan); err == nil {
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
