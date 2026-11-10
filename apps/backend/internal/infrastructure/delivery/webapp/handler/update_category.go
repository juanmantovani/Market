package handler

import (
	"net/http"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/categories"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/categories"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type UpdateCategoryHandler struct {
	updateCategory categories.UpdateCategory
}

func NewUpdateCategoryHandler(updateCategory categories.UpdateCategory) *UpdateCategoryHandler {
	return &UpdateCategoryHandler{updateCategory: updateCategory}
}

// Handle serves PATCH /api/v1/categories/{id}.
func (handler *UpdateCategoryHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	body, err := decodeJSON[dto.UpdateCategoryRequest](request)
	if err != nil {
		return err
	}

	category, err := handler.updateCategory.Execute(request.Context(), categories.UpdateCategoryInput{
		Actor:   actorFrom(request),
		ID:      request.PathValue("id"),
		Changes: domain.CategoryChanges{Name: body.Name, Description: body.Description},
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, dto.NewCategoryResponse(category))
	return nil
}
