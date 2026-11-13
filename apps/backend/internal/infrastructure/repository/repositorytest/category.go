// Package repositorytest holds the behavior every persistence adapter must
// share. Each contract runs against the in-memory adapter today and against
// PostgreSQL later, so swapping adapters is checked instead of assumed.
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

var contractTime = time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)

func text(value string) *string {
	return &value
}

func newCategory(t *testing.T, name string, description *string) domain.Category {
	t.Helper()
	category, err := domain.NewCategory(name, description, contractTime)
	if err != nil {
		t.Fatalf("domain.NewCategory(%q) error = %v", name, err)
	}
	return category
}

func mustCreate(t *testing.T, repository gateway.CategoryRepository, category domain.Category) domain.Category {
	t.Helper()
	created, err := repository.Create(context.Background(), category)
	if err != nil {
		t.Fatalf("Create(%q) error = %v", category.Name, err)
	}
	return created
}

func names(categories []domain.Category) []string {
	result := make([]string, len(categories))
	for i, category := range categories {
		result[i] = category.Name
	}
	return result
}

// CategoryRepository runs the gateway.CategoryRepository contract.
func CategoryRepository(t *testing.T, newRepositories NewRepositories) {
	ctx := context.Background()
	newRepository := func(t *testing.T) gateway.CategoryRepository {
		return newRepositories(t).Categories
	}

	t.Run("a category with items cannot be deleted", func(t *testing.T) {
		repositories := newRepositories(t)
		almacen := mustCreate(t, repositories.Categories, newCategory(t, "Almacén", nil))
		item := mustCreateItem(t, repositories.Items, newItem(t, almacen.ID, "YER-1", "Yerba"))

		if err := repositories.Categories.Delete(ctx, almacen.ID); !errors.Is(err, domain.ErrCategoryInUse) {
			t.Fatalf("Delete(category with items) error = %v, want ErrCategoryInUse", err)
		}
		if _, err := repositories.Categories.Get(ctx, almacen.ID); err != nil {
			t.Fatalf("the category is gone after a refused delete: %v", err)
		}

		if err := repositories.Items.Delete(ctx, item.ID); err != nil {
			t.Fatalf("Items.Delete() error = %v", err)
		}
		if err := repositories.Categories.Delete(ctx, almacen.ID); err != nil {
			t.Fatalf("Delete(category without items) error = %v", err)
		}
	})

	t.Run("create assigns an id and get returns the stored category", func(t *testing.T) {
		repository := newRepository(t)

		created := mustCreate(t, repository, newCategory(t, "Bebidas", text("Gaseosas y jugos")))
		if created.ID == "" {
			t.Fatal("Create did not assign an ID")
		}

		got, err := repository.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got.Name != "Bebidas" || got.Description == nil || *got.Description != "Gaseosas y jugos" {
			t.Fatalf("Get() = %+v", got)
		}
		if !got.CreatedAt.Equal(contractTime) || !got.UpdatedAt.Equal(contractTime) {
			t.Fatalf("timestamps = %v / %v, want %v", got.CreatedAt, got.UpdatedAt, contractTime)
		}
	})

	t.Run("create assigns distinct ids", func(t *testing.T) {
		repository := newRepository(t)

		first := mustCreate(t, repository, newCategory(t, "Bebidas", nil))
		second := mustCreate(t, repository, newCategory(t, "Almacén", nil))
		if first.ID == second.ID {
			t.Fatalf("both categories got ID %q", first.ID)
		}
	})

	t.Run("a name is unique ignoring case, accents and spacing", func(t *testing.T) {
		repository := newRepository(t)
		mustCreate(t, repository, newCategory(t, "Lácteos", nil))

		_, err := repository.Create(ctx, newCategory(t, "  LACTEOS ", nil))
		if !errors.Is(err, domain.ErrCategoryNameInUse) {
			t.Fatalf("Create(duplicate) error = %v, want ErrCategoryNameInUse", err)
		}
	})

	t.Run("delete removes the category and frees its name", func(t *testing.T) {
		repository := newRepository(t)
		kiosco := mustCreate(t, repository, newCategory(t, "Kiosco", nil))

		if err := repository.Delete(ctx, kiosco.ID); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		if _, err := repository.Get(ctx, kiosco.ID); !errors.Is(err, domain.ErrCategoryNotFound) {
			t.Fatalf("Get(deleted) error = %v, want ErrCategoryNotFound", err)
		}
		if err := repository.Delete(ctx, kiosco.ID); !errors.Is(err, domain.ErrCategoryNotFound) {
			t.Fatalf("Delete(deleted) error = %v, want ErrCategoryNotFound", err)
		}
		mustCreate(t, repository, newCategory(t, "kiosco", nil))
	})

	t.Run("update persists changes and keeps the name check", func(t *testing.T) {
		repository := newRepository(t)
		bebidas := mustCreate(t, repository, newCategory(t, "Bebidas", text("Gaseosas")))
		mustCreate(t, repository, newCategory(t, "Almacén", nil))

		later := contractTime.Add(time.Hour)
		changed, err := bebidas.Apply(domain.CategoryChanges{Name: text("Bebidas frías"), Description: text("")}, later)
		if err != nil {
			t.Fatalf("Apply() error = %v", err)
		}
		if _, err := repository.Update(ctx, changed); err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		got, _ := repository.Get(ctx, bebidas.ID)
		if got.Name != "Bebidas frías" || got.Description != nil || !got.UpdatedAt.Equal(later) {
			t.Fatalf("Get() after update = %+v", got)
		}

		renamed, _ := got.Apply(domain.CategoryChanges{Name: text("ALMACEN")}, later)
		if _, err := repository.Update(ctx, renamed); !errors.Is(err, domain.ErrCategoryNameInUse) {
			t.Fatalf("Update(into a taken name) error = %v, want ErrCategoryNameInUse", err)
		}

		unchanged, _ := repository.Update(ctx, got)
		if unchanged.Name != "Bebidas frías" {
			t.Fatalf("Update with its own name = %+v", unchanged)
		}
	})

	t.Run("unknown ids are not found", func(t *testing.T) {
		repository := newRepository(t)

		if _, err := repository.Get(ctx, "00000000-0000-7000-8000-000000000000"); !errors.Is(err, domain.ErrCategoryNotFound) {
			t.Fatalf("Get(unknown) error = %v, want ErrCategoryNotFound", err)
		}
		if _, err := repository.Get(ctx, "not-a-uuid"); !errors.Is(err, domain.ErrCategoryNotFound) {
			t.Fatalf("Get(malformed) error = %v, want ErrCategoryNotFound", err)
		}
		ghost := newCategory(t, "Fantasma", nil)
		ghost.ID = "00000000-0000-7000-8000-000000000000"
		if _, err := repository.Update(ctx, ghost); !errors.Is(err, domain.ErrCategoryNotFound) {
			t.Fatalf("Update(unknown) error = %v, want ErrCategoryNotFound", err)
		}
		if err := repository.Delete(ctx, ghost.ID); !errors.Is(err, domain.ErrCategoryNotFound) {
			t.Fatalf("Delete(unknown) error = %v, want ErrCategoryNotFound", err)
		}
	})

	t.Run("list searches and sorts by name", func(t *testing.T) {
		repository := newRepository(t)
		mustCreate(t, repository, newCategory(t, "Perfumería", nil))
		mustCreate(t, repository, newCategory(t, "almacén", text("Fideos y arroz")))
		mustCreate(t, repository, newCategory(t, "Bebidas", text("Gaseosas")))
		page := domain.PageRequest{Page: 1, Limit: 10}

		tests := []struct {
			name   string
			filter gateway.CategoryFilter
			want   []string
		}{
			{name: "all, sorted ignoring case and accents", filter: gateway.CategoryFilter{Page: page}, want: []string{"almacén", "Bebidas", "Perfumería"}},
			{name: "search by name without accents", filter: gateway.CategoryFilter{Search: "PERFUMERIA", Page: page}, want: []string{"Perfumería"}},
			{name: "search by description", filter: gateway.CategoryFilter{Search: "arroz", Page: page}, want: []string{"almacén"}},
			{name: "no match", filter: gateway.CategoryFilter{Search: "ferretería", Page: page}, want: []string{}},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				got, err := repository.List(ctx, test.filter)
				if err != nil {
					t.Fatalf("List() error = %v", err)
				}
				if fmt.Sprint(names(got.Items)) != fmt.Sprint(test.want) || got.Total != len(test.want) {
					t.Fatalf("List() = %v (total %d), want %v", names(got.Items), got.Total, test.want)
				}
				if got.Items == nil {
					t.Fatal("List() returned nil items, want an empty slice")
				}
			})
		}
	})

	t.Run("list pages through the results", func(t *testing.T) {
		repository := newRepository(t)
		for _, name := range []string{"E", "A", "D", "C", "B"} {
			mustCreate(t, repository, newCategory(t, name, nil))
		}

		second, err := repository.List(ctx, gateway.CategoryFilter{Page: domain.PageRequest{Page: 2, Limit: 2}})
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if fmt.Sprint(names(second.Items)) != "[C D]" || second.Total != 5 || second.Page != 2 || second.Limit != 2 {
			t.Fatalf("page 2 = %v %+v", names(second.Items), second)
		}

		beyond, _ := repository.List(ctx, gateway.CategoryFilter{Page: domain.PageRequest{Page: 9, Limit: 2}})
		if len(beyond.Items) != 0 || beyond.Total != 5 {
			t.Fatalf("page beyond the end = %v (total %d)", names(beyond.Items), beyond.Total)
		}
	})

	t.Run("callers cannot mutate stored categories", func(t *testing.T) {
		repository := newRepository(t)
		description := "Original"
		input := newCategory(t, "Bebidas", &description)
		created := mustCreate(t, repository, input)

		*created.Description = "changed through the result"
		got, _ := repository.Get(ctx, created.ID)
		*got.Description = "changed through get"

		again, _ := repository.Get(ctx, created.ID)
		if *again.Description != "Original" {
			t.Fatalf("stored description = %q, want Original", *again.Description)
		}
	})

	t.Run("concurrent creates of the same name keep only one", func(t *testing.T) {
		repository := newRepository(t)
		panaderia := newCategory(t, "Panadería", nil)
		const attempts = 20
		var wg sync.WaitGroup
		var mu sync.Mutex
		succeeded := 0

		for range attempts {
			wg.Go(func() {
				if _, err := repository.Create(ctx, panaderia); err == nil {
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
