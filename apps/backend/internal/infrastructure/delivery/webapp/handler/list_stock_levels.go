package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/stock"
	pagination "market/backend/internal/infrastructure/delivery/webapp/dto/pagination"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/stock"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type ListStockLevelsHandler struct {
	listStockLevels stock.ListStockLevels
}

func NewListStockLevelsHandler(listStockLevels stock.ListStockLevels) *ListStockLevelsHandler {
	return &ListStockLevelsHandler{listStockLevels: listStockLevels}
}

// Handle serves GET /api/v1/stock?q=&categoryId=&lowStock=true&page=&limit=.
// Any lowStock other than "true" lists every item.
func (handler *ListStockLevelsHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	page, limit, err := pageQuery(request)
	if err != nil {
		return err
	}

	query := request.URL.Query()
	result, err := handler.listStockLevels.Execute(request.Context(), stock.ListStockLevelsInput{
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

	response.WriteJSON(w, http.StatusOK, pagination.NewPageResponse(result, dto.NewStockLevelResponse))
	return nil
}
