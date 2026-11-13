package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/items"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/items"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type GetItemHandler struct {
	getItem items.GetItem
}

func NewGetItemHandler(getItem items.GetItem) *GetItemHandler {
	return &GetItemHandler{getItem: getItem}
}

// Handle serves GET /api/v1/items/{id}.
func (handler *GetItemHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	item, err := handler.getItem.Execute(request.Context(), items.GetItemInput{
		Actor: actorFrom(request),
		ID:    request.PathValue("id"),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, dto.NewItemResponse(item))
	return nil
}
