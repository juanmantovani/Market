package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/categories"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/categories"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type DeactivateCategoryHandler struct {
	deactivateCategory categories.DeactivateCategory
}

func NewDeactivateCategoryHandler(deactivateCategory categories.DeactivateCategory) *DeactivateCategoryHandler {
	return &DeactivateCategoryHandler{deactivateCategory: deactivateCategory}
}

// Handle serves DELETE /api/v1/categories/{id}. The deletion is logical, so
// it answers with the deactivated category.
func (handler *DeactivateCategoryHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	category, err := handler.deactivateCategory.Execute(request.Context(), categories.DeactivateCategoryInput{
		Actor: actorFrom(request),
		ID:    request.PathValue("id"),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, dto.NewCategoryResponse(category))
	return nil
}
