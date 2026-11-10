package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/categories"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/categories"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type ReactivateCategoryHandler struct {
	reactivateCategory categories.ReactivateCategory
}

func NewReactivateCategoryHandler(reactivateCategory categories.ReactivateCategory) *ReactivateCategoryHandler {
	return &ReactivateCategoryHandler{reactivateCategory: reactivateCategory}
}

// Handle serves POST /api/v1/categories/{id}/reactivate.
func (handler *ReactivateCategoryHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	category, err := handler.reactivateCategory.Execute(request.Context(), categories.ReactivateCategoryInput{
		Actor: actorFrom(request),
		ID:    request.PathValue("id"),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, dto.NewCategoryResponse(category))
	return nil
}
