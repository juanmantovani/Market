package handler

import (
	"net/http"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/stock"
	pagination "market/backend/internal/infrastructure/delivery/webapp/dto/pagination"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/stock"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type ListStockMovementsHandler struct {
	listStockMovements stock.ListStockMovements
}

func NewListStockMovementsHandler(listStockMovements stock.ListStockMovements) *ListStockMovementsHandler {
	return &ListStockMovementsHandler{listStockMovements: listStockMovements}
}

// Handle serves GET /api/v1/stock/movements?itemId=&type=&page=&limit=.
func (handler *ListStockMovementsHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	page, limit, err := pageQuery(request)
	if err != nil {
		return err
	}

	query := request.URL.Query()
	result, err := handler.listStockMovements.Execute(request.Context(), stock.ListStockMovementsInput{
		Actor:  actorFrom(request),
		ItemID: query.Get("itemId"),
		Type:   domain.MovementType(query.Get("type")),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, pagination.NewPageResponse(result, dto.NewStockMovementResponse))
	return nil
}
