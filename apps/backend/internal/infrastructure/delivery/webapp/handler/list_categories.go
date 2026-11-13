package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/categories"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/categories"
	pagination "market/backend/internal/infrastructure/delivery/webapp/dto/pagination"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type ListCategoriesHandler struct {
	listCategories categories.ListCategories
}

func NewListCategoriesHandler(listCategories categories.ListCategories) *ListCategoriesHandler {
	return &ListCategoriesHandler{listCategories: listCategories}
}

// Handle serves GET /api/v1/categories?q=&page=&limit=.
func (handler *ListCategoriesHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	page, limit, err := pageQuery(request)
	if err != nil {
		return err
	}

	query := request.URL.Query()
	result, err := handler.listCategories.Execute(request.Context(), categories.ListCategoriesInput{
		Actor:  actorFrom(request),
		Search: query.Get("q"),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, pagination.NewPageResponse(result, dto.NewCategoryResponse))
	return nil
}
