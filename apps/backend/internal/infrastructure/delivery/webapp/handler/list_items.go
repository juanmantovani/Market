package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/items"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/items"
	pagination "market/backend/internal/infrastructure/delivery/webapp/dto/pagination"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type ListItemsHandler struct {
	listItems items.ListItems
}

func NewListItemsHandler(listItems items.ListItems) *ListItemsHandler {
	return &ListItemsHandler{listItems: listItems}
}

// Handle serves GET /api/v1/items?q=&categoryId=&lowStock=true&page=&limit=.
func (handler *ListItemsHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	page, limit, err := pageQuery(request)
	if err != nil {
		return err
	}

	query := request.URL.Query()
	result, err := handler.listItems.Execute(request.Context(), items.ListItemsInput{
		Actor:      actorFrom(request),
		Search:     query.Get("q"),
		CategoryID: query.Get("categoryId"),
		LowStock:   query.Get("lowStock") == "true",
		Page:       page,
		Limit:      limit,
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, pagination.NewPageResponse(result, dto.NewItemResponse))
	return nil
}
