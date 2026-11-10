package categories

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
	now     = time.Date(2026, 11, 10, 15, 0, 0, 0, time.UTC)
	clock   = gatewayfake.Clock{Time: now}
	admin   = domain.Actor{ID: "u-admin", Role: domain.RoleAdmin}
	manager = domain.Actor{ID: "u-manager", Role: domain.RoleManager}
	cashier = domain.Actor{ID: "u-cashier", Role: domain.RoleCashier}
	nobody  = domain.Actor{}

	errDown = errors.New("connection refused")
)

func ptr[T any](value T) *T {
	return &value
}

func storedCategory(active bool) domain.Category {
	return domain.Category{
		ID:        "category-1",
		Name:      "Bebidas",
		Active:    active,
		CreatedAt: now.Add(-time.Hour),
		UpdatedAt: now.Add(-time.Hour),
	}
}

func assertError(t *testing.T, err error, want *domain.Error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %s", err, want.Code)
	}
}

func TestCreateCategory(t *testing.T) {
	repository := &gatewayfake.CategoryRepository{}
	useCase := NewCreateCategory(repository, clock)

	created, err := useCase.Execute(context.Background(), CreateCategoryInput{
		Actor: manager, Name: " Bebidas ", Description: ptr("Gaseosas"),
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if created.ID == "" || created.Name != "Bebidas" || !created.Active || !created.CreatedAt.Equal(now) {
		t.Fatalf("created = %+v", created)
	}
	if len(repository.Created) != 1 {
		t.Fatalf("repository received %d creates, want 1", len(repository.Created))
	}
}

func TestCreateCategoryRejectsBeforeStoring(t *testing.T) {
	tests := []struct {
		name  string
		input CreateCategoryInput
		want  *domain.Error
	}{
		{name: "cashier", input: CreateCategoryInput{Actor: cashier, Name: "Bebidas"}, want: domain.ErrForbidden},
		{name: "no actor", input: CreateCategoryInput{Actor: nobody, Name: "Bebidas"}, want: domain.ErrForbidden},
		{name: "blank name", input: CreateCategoryInput{Actor: admin, Name: " "}, want: domain.ErrCategoryNameRequired},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &gatewayfake.CategoryRepository{}

			_, err := NewCreateCategory(repository, clock).Execute(context.Background(), test.input)

			assertError(t, err, test.want)
			if len(repository.Created) != 0 {
				t.Fatal("repository was called")
			}
		})
	}
}

func TestCreateCategoryTranslatesRepositoryErrors(t *testing.T) {
	conflict := &gatewayfake.CategoryRepository{CreateErr: domain.ErrCategoryNameInUse}
	_, err := NewCreateCategory(conflict, clock).Execute(context.Background(), CreateCategoryInput{Actor: admin, Name: "Bebidas"})
	assertError(t, err, domain.ErrCategoryNameInUse)

	down := &gatewayfake.CategoryRepository{CreateErr: errDown}
	_, err = NewCreateCategory(down, clock).Execute(context.Background(), CreateCategoryInput{Actor: admin, Name: "Bebidas"})
	assertError(t, err, domain.ErrStorageUnavailable)
	if !errors.Is(err, errDown) {
		t.Fatal("storage error lost its cause")
	}
}

func TestListCategories(t *testing.T) {
	page := domain.Page[domain.Category]{Items: []domain.Category{storedCategory(true)}, Total: 1, Page: 2, Limit: 10}
	repository := &gatewayfake.CategoryRepository{ListResult: page}

	got, err := NewListCategories(repository).Execute(context.Background(), ListCategoriesInput{
		Actor: cashier, Search: "  beb ", Status: "all", Page: 2, Limit: 10,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got.Total != 1 || len(got.Items) != 1 {
		t.Fatalf("page = %+v", got)
	}
	want := gateway.CategoryFilter{Search: "beb", Status: domain.StatusAll, Page: domain.PageRequest{Page: 2, Limit: 10}}
	if repository.ListFilters[0] != want {
		t.Fatalf("filter = %+v, want %+v", repository.ListFilters[0], want)
	}
}

func TestListCategoriesDefaultsToActiveFirstPage(t *testing.T) {
	repository := &gatewayfake.CategoryRepository{}

	if _, err := NewListCategories(repository).Execute(context.Background(), ListCategoriesInput{Actor: cashier}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	want := gateway.CategoryFilter{Status: domain.StatusActive, Page: domain.PageRequest{Page: 1, Limit: domain.DefaultPageLimit}}
	if repository.ListFilters[0] != want {
		t.Fatalf("filter = %+v, want %+v", repository.ListFilters[0], want)
	}
}

func TestListCategoriesRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name  string
		input ListCategoriesInput
		want  *domain.Error
	}{
		{name: "no actor", input: ListCategoriesInput{}, want: domain.ErrForbidden},
		{name: "status", input: ListCategoriesInput{Actor: admin, Status: "deleted"}, want: domain.ErrInvalidStatusFilter},
		{name: "limit", input: ListCategoriesInput{Actor: admin, Limit: 500}, want: domain.ErrInvalidPage},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &gatewayfake.CategoryRepository{}

			_, err := NewListCategories(repository).Execute(context.Background(), test.input)

			assertError(t, err, test.want)
			if len(repository.ListFilters) != 0 {
				t.Fatal("repository was called")
			}
		})
	}
}

func TestListCategoriesTranslatesStorageErrors(t *testing.T) {
	repository := &gatewayfake.CategoryRepository{ListErr: errDown}

	_, err := NewListCategories(repository).Execute(context.Background(), ListCategoriesInput{Actor: admin})

	assertError(t, err, domain.ErrStorageUnavailable)
}

func TestGetCategory(t *testing.T) {
	repository := &gatewayfake.CategoryRepository{GetResult: storedCategory(true)}

	got, err := NewGetCategory(repository).Execute(context.Background(), GetCategoryInput{Actor: cashier, ID: "category-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got.ID != "category-1" || repository.GetIDs[0] != "category-1" {
		t.Fatalf("got %+v from IDs %v", got, repository.GetIDs)
	}

	_, err = NewGetCategory(repository).Execute(context.Background(), GetCategoryInput{Actor: nobody, ID: "category-1"})
	assertError(t, err, domain.ErrForbidden)

	missing := &gatewayfake.CategoryRepository{GetErr: domain.ErrCategoryNotFound}
	_, err = NewGetCategory(missing).Execute(context.Background(), GetCategoryInput{Actor: admin, ID: "x"})
	assertError(t, err, domain.ErrCategoryNotFound)
}

func TestUpdateCategory(t *testing.T) {
	repository := &gatewayfake.CategoryRepository{GetResult: storedCategory(true)}

	updated, err := NewUpdateCategory(repository, clock).Execute(context.Background(), UpdateCategoryInput{
		Actor: admin, ID: "category-1", Changes: domain.CategoryChanges{Name: ptr("Bebidas frías")},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if updated.Name != "Bebidas frías" || !updated.UpdatedAt.Equal(now) {
		t.Fatalf("updated = %+v", updated)
	}
	if len(repository.Updated) != 1 {
		t.Fatalf("repository received %d updates, want 1", len(repository.Updated))
	}
}

func TestUpdateCategoryRejectsBeforeStoring(t *testing.T) {
	tests := []struct {
		name       string
		repository *gatewayfake.CategoryRepository
		input      UpdateCategoryInput
		want       *domain.Error
	}{
		{
			name:       "cashier",
			repository: &gatewayfake.CategoryRepository{GetResult: storedCategory(true)},
			input:      UpdateCategoryInput{Actor: cashier, ID: "category-1", Changes: domain.CategoryChanges{Name: ptr("X")}},
			want:       domain.ErrForbidden,
		},
		{
			name:       "no changes",
			repository: &gatewayfake.CategoryRepository{GetResult: storedCategory(true)},
			input:      UpdateCategoryInput{Actor: admin, ID: "category-1"},
			want:       domain.ErrCategoryUpdateEmpty,
		},
		{
			name:       "not found",
			repository: &gatewayfake.CategoryRepository{GetErr: domain.ErrCategoryNotFound},
			input:      UpdateCategoryInput{Actor: admin, ID: "nope", Changes: domain.CategoryChanges{Name: ptr("X")}},
			want:       domain.ErrCategoryNotFound,
		},
		{
			name:       "invalid name",
			repository: &gatewayfake.CategoryRepository{GetResult: storedCategory(true)},
			input:      UpdateCategoryInput{Actor: admin, ID: "category-1", Changes: domain.CategoryChanges{Name: ptr("")}},
			want:       domain.ErrCategoryNameRequired,
		},
		{
			name:       "storage down on read",
			repository: &gatewayfake.CategoryRepository{GetErr: errDown},
			input:      UpdateCategoryInput{Actor: admin, ID: "category-1", Changes: domain.CategoryChanges{Name: ptr("X")}},
			want:       domain.ErrStorageUnavailable,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewUpdateCategory(test.repository, clock).Execute(context.Background(), test.input)

			assertError(t, err, test.want)
			if len(test.repository.Updated) != 0 {
				t.Fatal("repository stored an update")
			}
		})
	}
}

func TestUpdateCategoryTranslatesWriteErrors(t *testing.T) {
	conflict := &gatewayfake.CategoryRepository{GetResult: storedCategory(true), UpdateErr: domain.ErrCategoryNameInUse}
	input := UpdateCategoryInput{Actor: admin, ID: "category-1", Changes: domain.CategoryChanges{Name: ptr("Almacén")}}

	_, err := NewUpdateCategory(conflict, clock).Execute(context.Background(), input)
	assertError(t, err, domain.ErrCategoryNameInUse)

	down := &gatewayfake.CategoryRepository{GetResult: storedCategory(true), UpdateErr: errDown}
	_, err = NewUpdateCategory(down, clock).Execute(context.Background(), input)
	assertError(t, err, domain.ErrStorageUnavailable)
}

func TestDeactivateCategory(t *testing.T) {
	repository := &gatewayfake.CategoryRepository{GetResult: storedCategory(true)}

	got, err := NewDeactivateCategory(repository, clock).Execute(context.Background(), DeactivateCategoryInput{Actor: manager, ID: "category-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got.Active || !got.UpdatedAt.Equal(now) || len(repository.Updated) != 1 {
		t.Fatalf("got %+v after %d updates", got, len(repository.Updated))
	}
}

func TestReactivateCategory(t *testing.T) {
	repository := &gatewayfake.CategoryRepository{GetResult: storedCategory(false)}

	got, err := NewReactivateCategory(repository, clock).Execute(context.Background(), ReactivateCategoryInput{Actor: admin, ID: "category-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !got.Active || len(repository.Updated) != 1 {
		t.Fatalf("got %+v after %d updates", got, len(repository.Updated))
	}
}

func TestSetActiveIsIdempotent(t *testing.T) {
	alreadyInactive := &gatewayfake.CategoryRepository{GetResult: storedCategory(false)}
	got, err := NewDeactivateCategory(alreadyInactive, clock).Execute(context.Background(), DeactivateCategoryInput{Actor: admin, ID: "category-1"})
	if err != nil || got.Active || len(alreadyInactive.Updated) != 0 {
		t.Fatalf("deactivating an inactive category: %+v, %v, %d updates", got, err, len(alreadyInactive.Updated))
	}

	alreadyActive := &gatewayfake.CategoryRepository{GetResult: storedCategory(true)}
	got, err = NewReactivateCategory(alreadyActive, clock).Execute(context.Background(), ReactivateCategoryInput{Actor: admin, ID: "category-1"})
	if err != nil || !got.Active || len(alreadyActive.Updated) != 0 {
		t.Fatalf("reactivating an active category: %+v, %v, %d updates", got, err, len(alreadyActive.Updated))
	}
}

func TestSetActiveErrors(t *testing.T) {
	_, err := NewDeactivateCategory(&gatewayfake.CategoryRepository{}, clock).Execute(context.Background(), DeactivateCategoryInput{Actor: cashier, ID: "category-1"})
	assertError(t, err, domain.ErrForbidden)

	missing := &gatewayfake.CategoryRepository{GetErr: domain.ErrCategoryNotFound}
	_, err = NewDeactivateCategory(missing, clock).Execute(context.Background(), DeactivateCategoryInput{Actor: admin, ID: "x"})
	assertError(t, err, domain.ErrCategoryNotFound)

	nameTaken := &gatewayfake.CategoryRepository{GetResult: storedCategory(false), UpdateErr: domain.ErrCategoryNameInUse}
	_, err = NewReactivateCategory(nameTaken, clock).Execute(context.Background(), ReactivateCategoryInput{Actor: admin, ID: "category-1"})
	assertError(t, err, domain.ErrCategoryNameInUse)
}
