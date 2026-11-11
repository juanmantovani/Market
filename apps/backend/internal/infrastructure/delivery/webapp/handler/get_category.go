package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/categories"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/categories"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type GetCategoryHandler struct {
	getCategory categories.GetCategory
}

func NewGetCategoryHandler(getCategory categories.GetCategory) *GetCategoryHandler {
	return &GetCategoryHandler{getCategory: getCategory}
}

// Handle serves GET /api/v1/categories/{id}.
func (handler *GetCategoryHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	category, err := handler.getCategory.Execute(request.Context(), categories.GetCategoryInput{
		Actor: actorFrom(request),
		ID:    request.PathValue("id"),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, dto.NewCategoryResponse(category))
	return nil
}
