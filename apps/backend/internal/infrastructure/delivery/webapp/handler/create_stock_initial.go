package handler

import (
	"net/http"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/stock"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/stock"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type CreateStockInitialHandler struct {
	registerInitialStock stock.RegisterInitialStock
}

func NewCreateStockInitialHandler(registerInitialStock stock.RegisterInitialStock) *CreateStockInitialHandler {
	return &CreateStockInitialHandler{registerInitialStock: registerInitialStock}
}

// Handle serves POST /api/v1/stock/initial.
func (handler *CreateStockInitialHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	body, err := decodeJSON[dto.StockInitialRequest](request)
	if err != nil {
		return err
	}

	movement, err := handler.registerInitialStock.Execute(request.Context(), stock.RegisterInitialStockInput{
		Actor:    actorFrom(request),
		ItemID:   body.ItemID,
		Quantity: domain.Quantity(body.Quantity),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusCreated, dto.NewStockMovementResponse(movement))
	return nil
}
