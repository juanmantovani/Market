package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/categories"
)

type DeleteCategoryHandler struct {
	deleteCategory categories.DeleteCategory
}

func NewDeleteCategoryHandler(deleteCategory categories.DeleteCategory) *DeleteCategoryHandler {
	return &DeleteCategoryHandler{deleteCategory: deleteCategory}
}

// Handle serves DELETE /api/v1/categories/{id}.
func (handler *DeleteCategoryHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	err := handler.deleteCategory.Execute(request.Context(), categories.DeleteCategoryInput{
		Actor: actorFrom(request),
		ID:    request.PathValue("id"),
	})
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
