package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/items"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/items"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type CreateItemHandler struct {
	createItem items.CreateItem
}

func NewCreateItemHandler(createItem items.CreateItem) *CreateItemHandler {
	return &CreateItemHandler{createItem: createItem}
}

// Handle serves POST /api/v1/items.
func (handler *CreateItemHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	body, err := decodeJSON[dto.ItemRequest](request)
	if err != nil {
		return err
	}

	item, err := handler.createItem.Execute(request.Context(), items.CreateItemInput{
		Actor:  actorFrom(request),
		Fields: body.Fields(),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusCreated, dto.NewItemResponse(item))
	return nil
}
