package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/items"
)

type DeleteItemHandler struct {
	deleteItem items.DeleteItem
}

func NewDeleteItemHandler(deleteItem items.DeleteItem) *DeleteItemHandler {
	return &DeleteItemHandler{deleteItem: deleteItem}
}

// Handle serves DELETE /api/v1/items/{id}.
func (handler *DeleteItemHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	err := handler.deleteItem.Execute(request.Context(), items.DeleteItemInput{
		Actor: actorFrom(request),
		ID:    request.PathValue("id"),
	})
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
