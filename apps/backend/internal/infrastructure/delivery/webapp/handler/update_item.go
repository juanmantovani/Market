package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/items"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/items"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type UpdateItemHandler struct {
	updateItem items.UpdateItem
}

func NewUpdateItemHandler(updateItem items.UpdateItem) *UpdateItemHandler {
	return &UpdateItemHandler{updateItem: updateItem}
}

// Handle serves PUT /api/v1/items/{id}: the body replaces every editable
// field, so an optional field left out is cleared.
func (handler *UpdateItemHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	body, err := decodeJSON[dto.ItemRequest](request)
	if err != nil {
		return err
	}

	item, err := handler.updateItem.Execute(request.Context(), items.UpdateItemInput{
		Actor:  actorFrom(request),
		ID:     request.PathValue("id"),
		Fields: body.Fields(),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, dto.NewItemResponse(item))
	return nil
}
