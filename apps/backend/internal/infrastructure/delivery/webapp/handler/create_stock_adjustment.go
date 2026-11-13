package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/stock"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/stock"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type CreateStockAdjustmentHandler struct {
	registerStockAdjustment stock.RegisterStockAdjustment
}

func NewCreateStockAdjustmentHandler(registerStockAdjustment stock.RegisterStockAdjustment) *CreateStockAdjustmentHandler {
	return &CreateStockAdjustmentHandler{registerStockAdjustment: registerStockAdjustment}
}

// Handle serves POST /api/v1/stock/adjustments.
func (handler *CreateStockAdjustmentHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	body, err := decodeJSON[dto.StockAdjustmentRequest](request)
	if err != nil {
		return err
	}

	movement, err := handler.registerStockAdjustment.Execute(request.Context(), stock.RegisterStockAdjustmentInput{
		Actor:      actorFrom(request),
		Adjustment: body.Adjustment(),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusCreated, dto.NewStockMovementResponse(movement))
	return nil
}
