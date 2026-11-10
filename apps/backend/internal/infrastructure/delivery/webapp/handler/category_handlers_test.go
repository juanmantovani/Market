package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/categories"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/categories"
	pagination "market/backend/internal/infrastructure/delivery/webapp/dto/pagination"
	"market/backend/internal/infrastructure/delivery/webapp/middleware"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

var (
	handlerActor    = domain.Actor{ID: "u1", Name: "Ana", Role: domain.RoleManager}
	handlerTime     = time.Date(2026, 11, 10, 15, 0, 0, 0, time.UTC)
	handlerCategory = domain.Category{
		ID: "c1", Name: "Bebidas", Description: ptr("Gaseosas"), Active: true,
		CreatedAt: handlerTime, UpdatedAt: handlerTime,
	}
)

func ptr[T any](value T) *T {
	return &value
}

type useCaseFunc[I, O any] func(ctx context.Context, input I) (O, error)

func (fn useCaseFunc[I, O]) Execute(ctx context.Context, input I) (O, error) {
	return fn(ctx, input)
}

func categoryRequest(method, target string, body io.Reader) *http.Request {
	request := httptest.NewRequest(method, target, body)
	return request.WithContext(middleware.WithActor(request.Context(), handlerActor))
}

func serve(h HTTPHandler, request *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	Adapt(h, time.Second)(recorder, request)
	return recorder
}

func assertStatus(t *testing.T, recorder *httptest.ResponseRecorder, want int) {
	t.Helper()
	if recorder.Code != want {
		t.Fatalf("status = %d, want %d (body %s)", recorder.Code, want, recorder.Body.String())
	}
}

func assertCategoryBody(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()
	got := decodeBody[dto.CategoryResponse](t, recorder)
	if got.ID != "c1" || got.Name != "Bebidas" || got.Description == nil || *got.Description != "Gaseosas" || !got.Active || !got.CreatedAt.Equal(handlerTime) {
		t.Fatalf("body = %+v", got)
	}
}

func TestListCategoriesHandler(t *testing.T) {
	var received categories.ListCategoriesInput
	listCategories := useCaseFunc[categories.ListCategoriesInput, domain.Page[domain.Category]](
		func(_ context.Context, input categories.ListCategoriesInput) (domain.Page[domain.Category], error) {
			received = input
			return domain.Page[domain.Category]{Items: []domain.Category{handlerCategory}, Total: 31, Page: 2, Limit: 10}, nil
		})

	recorder := serve(NewListCategoriesHandler(listCategories), categoryRequest(http.MethodGet, "/api/v1/categories?q=beb&status=all&page=2&limit=10", nil))

	assertStatus(t, recorder, http.StatusOK)
	got := decodeBody[pagination.PageResponse[dto.CategoryResponse]](t, recorder)
	if got.Total != 31 || got.Page != 2 || got.Limit != 10 || len(got.Items) != 1 || got.Items[0].Name != "Bebidas" {
		t.Fatalf("body = %+v", got)
	}
	want := categories.ListCategoriesInput{Actor: handlerActor, Search: "beb", Status: "all", Page: 2, Limit: 10}
	if received != want {
		t.Fatalf("input = %+v, want %+v", received, want)
	}
}

func TestListCategoriesHandlerAnswersAnEmptyArray(t *testing.T) {
	listCategories := useCaseFunc[categories.ListCategoriesInput, domain.Page[domain.Category]](
		func(context.Context, categories.ListCategoriesInput) (domain.Page[domain.Category], error) {
			return domain.Page[domain.Category]{Page: 1, Limit: 25}, nil
		})

	recorder := serve(NewListCategoriesHandler(listCategories), categoryRequest(http.MethodGet, "/api/v1/categories", nil))

	assertStatus(t, recorder, http.StatusOK)
	if !strings.Contains(recorder.Body.String(), `"items":[]`) {
		t.Fatalf("body = %s, want an empty items array", recorder.Body.String())
	}
}

func TestListCategoriesHandlerRejectsANonNumericPage(t *testing.T) {
	called := false
	listCategories := useCaseFunc[categories.ListCategoriesInput, domain.Page[domain.Category]](
		func(context.Context, categories.ListCategoriesInput) (domain.Page[domain.Category], error) {
			called = true
			return domain.Page[domain.Category]{}, nil
		})

	for _, target := range []string{"/api/v1/categories?page=two", "/api/v1/categories?limit=x"} {
		recorder := serve(NewListCategoriesHandler(listCategories), categoryRequest(http.MethodGet, target, nil))

		assertStatus(t, recorder, http.StatusBadRequest)
		if got := decodeBody[response.ErrorResponse](t, recorder); got.Code != "invalid_page" {
			t.Fatalf("%s: code = %q, want invalid_page", target, got.Code)
		}
	}
	if called {
		t.Fatal("use case ran with an invalid page")
	}
}

func TestGetCategoryHandler(t *testing.T) {
	var received categories.GetCategoryInput
	getCategory := useCaseFunc[categories.GetCategoryInput, domain.Category](
		func(_ context.Context, input categories.GetCategoryInput) (domain.Category, error) {
			received = input
			return handlerCategory, nil
		})
	request := categoryRequest(http.MethodGet, "/api/v1/categories/c1", nil)
	request.SetPathValue("id", "c1")

	recorder := serve(NewGetCategoryHandler(getCategory), request)

	assertStatus(t, recorder, http.StatusOK)
	assertCategoryBody(t, recorder)
	if received.ID != "c1" || received.Actor != handlerActor {
		t.Fatalf("input = %+v", received)
	}
}

func TestCreateCategoryHandler(t *testing.T) {
	var received categories.CreateCategoryInput
	createCategory := useCaseFunc[categories.CreateCategoryInput, domain.Category](
		func(_ context.Context, input categories.CreateCategoryInput) (domain.Category, error) {
			received = input
			return handlerCategory, nil
		})
	body := strings.NewReader(`{"name":"Bebidas","description":"Gaseosas"}`)

	recorder := serve(NewCreateCategoryHandler(createCategory), categoryRequest(http.MethodPost, "/api/v1/categories", body))

	assertStatus(t, recorder, http.StatusCreated)
	assertCategoryBody(t, recorder)
	if received.Name != "Bebidas" || received.Description == nil || *received.Description != "Gaseosas" || received.Actor != handlerActor {
		t.Fatalf("input = %+v", received)
	}
}

func TestCreateCategoryHandlerErrors(t *testing.T) {
	conflict := useCaseFunc[categories.CreateCategoryInput, domain.Category](
		func(context.Context, categories.CreateCategoryInput) (domain.Category, error) {
			return domain.Category{}, domain.ErrCategoryNameInUse
		})

	recorder := serve(NewCreateCategoryHandler(conflict), categoryRequest(http.MethodPost, "/api/v1/categories", strings.NewReader(`{"name":"Bebidas"}`)))
	assertStatus(t, recorder, http.StatusConflict)
	if got := decodeBody[response.ErrorResponse](t, recorder); got.Code != "category_name_in_use" {
		t.Fatalf("code = %q", got.Code)
	}

	recorder = serve(NewCreateCategoryHandler(conflict), categoryRequest(http.MethodPost, "/api/v1/categories", strings.NewReader(`{`)))
	assertStatus(t, recorder, http.StatusBadRequest)
}

func TestUpdateCategoryHandler(t *testing.T) {
	var received categories.UpdateCategoryInput
	updateCategory := useCaseFunc[categories.UpdateCategoryInput, domain.Category](
		func(_ context.Context, input categories.UpdateCategoryInput) (domain.Category, error) {
			received = input
			return handlerCategory, nil
		})
	request := categoryRequest(http.MethodPatch, "/api/v1/categories/c1", strings.NewReader(`{"description":""}`))
	request.SetPathValue("id", "c1")

	recorder := serve(NewUpdateCategoryHandler(updateCategory), request)

	assertStatus(t, recorder, http.StatusOK)
	assertCategoryBody(t, recorder)
	if received.ID != "c1" || received.Changes.Name != nil || received.Changes.Description == nil || *received.Changes.Description != "" {
		t.Fatalf("input = %+v, want only an empty description", received)
	}

	malformed := categoryRequest(http.MethodPatch, "/api/v1/categories/c1", strings.NewReader(`[]`))
	assertStatus(t, serve(NewUpdateCategoryHandler(updateCategory), malformed), http.StatusBadRequest)
}

func TestDeactivateAndReactivateCategoryHandlers(t *testing.T) {
	var deactivated, reactivated string
	deactivate := useCaseFunc[categories.DeactivateCategoryInput, domain.Category](
		func(_ context.Context, input categories.DeactivateCategoryInput) (domain.Category, error) {
			deactivated = input.ID
			return handlerCategory, nil
		})
	reactivate := useCaseFunc[categories.ReactivateCategoryInput, domain.Category](
		func(_ context.Context, input categories.ReactivateCategoryInput) (domain.Category, error) {
			reactivated = input.ID
			return handlerCategory, nil
		})

	request := categoryRequest(http.MethodDelete, "/api/v1/categories/c1", nil)
	request.SetPathValue("id", "c1")
	recorder := serve(NewDeactivateCategoryHandler(deactivate), request)
	assertStatus(t, recorder, http.StatusOK)
	assertCategoryBody(t, recorder)

	request = categoryRequest(http.MethodPost, "/api/v1/categories/c2/reactivate", nil)
	request.SetPathValue("id", "c2")
	recorder = serve(NewReactivateCategoryHandler(reactivate), request)
	assertStatus(t, recorder, http.StatusOK)

	if deactivated != "c1" || reactivated != "c2" {
		t.Fatalf("ids = %q / %q", deactivated, reactivated)
	}
}

func TestCategoryHandlersPassErrorsThrough(t *testing.T) {
	notFound := func(context.Context, categories.GetCategoryInput) (domain.Category, error) {
		return domain.Category{}, domain.ErrCategoryNotFound
	}
	forbidden := func(context.Context, categories.UpdateCategoryInput) (domain.Category, error) {
		return domain.Category{}, domain.ErrForbidden
	}
	deactivateFails := func(context.Context, categories.DeactivateCategoryInput) (domain.Category, error) {
		return domain.Category{}, domain.ErrCategoryNotFound
	}
	reactivateFails := func(context.Context, categories.ReactivateCategoryInput) (domain.Category, error) {
		return domain.Category{}, domain.ErrCategoryNameInUse
	}
	listFails := func(context.Context, categories.ListCategoriesInput) (domain.Page[domain.Category], error) {
		return domain.Page[domain.Category]{}, domain.ErrStorageUnavailable
	}

	tests := []struct {
		name    string
		handler HTTPHandler
		request *http.Request
		want    int
	}{
		{name: "get", handler: NewGetCategoryHandler(useCaseFunc[categories.GetCategoryInput, domain.Category](notFound)), request: categoryRequest(http.MethodGet, "/api/v1/categories/x", nil), want: http.StatusNotFound},
		{name: "update", handler: NewUpdateCategoryHandler(useCaseFunc[categories.UpdateCategoryInput, domain.Category](forbidden)), request: categoryRequest(http.MethodPatch, "/api/v1/categories/x", strings.NewReader(`{"name":"X"}`)), want: http.StatusForbidden},
		{name: "deactivate", handler: NewDeactivateCategoryHandler(useCaseFunc[categories.DeactivateCategoryInput, domain.Category](deactivateFails)), request: categoryRequest(http.MethodDelete, "/api/v1/categories/x", nil), want: http.StatusNotFound},
		{name: "reactivate", handler: NewReactivateCategoryHandler(useCaseFunc[categories.ReactivateCategoryInput, domain.Category](reactivateFails)), request: categoryRequest(http.MethodPost, "/api/v1/categories/x/reactivate", nil), want: http.StatusConflict},
		{name: "list", handler: NewListCategoriesHandler(useCaseFunc[categories.ListCategoriesInput, domain.Page[domain.Category]](listFails)), request: categoryRequest(http.MethodGet, "/api/v1/categories", nil), want: http.StatusServiceUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertStatus(t, serve(test.handler, test.request), test.want)
		})
	}
}

func TestActorFromARequestWithoutActor(t *testing.T) {
	if actor := actorFrom(httptest.NewRequest(http.MethodGet, "/", nil)); actor != (domain.Actor{}) {
		t.Fatalf("actorFrom() = %+v, want the zero actor", actor)
	}
}
