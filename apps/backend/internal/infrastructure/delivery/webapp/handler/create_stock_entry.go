package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/stock"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/stock"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type CreateStockEntryHandler struct {
	registerStockEntry stock.RegisterStockEntry
}

func NewCreateStockEntryHandler(registerStockEntry stock.RegisterStockEntry) *CreateStockEntryHandler {
	return &CreateStockEntryHandler{registerStockEntry: registerStockEntry}
}

// Handle serves POST /api/v1/stock/entries.
func (handler *CreateStockEntryHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	body, err := decodeJSON[dto.StockEntryRequest](request)
	if err != nil {
		return err
	}

	movements, err := handler.registerStockEntry.Execute(request.Context(), stock.RegisterStockEntryInput{
		Actor: actorFrom(request),
		Entry: body.Entry(),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusCreated, dto.NewStockEntryResponse(movements))
	return nil
}
