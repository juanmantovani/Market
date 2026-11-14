package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/items"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/items"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type LookupItemByBarcodeHandler struct {
	lookupItemByBarcode items.LookupItemByBarcode
}

func NewLookupItemByBarcodeHandler(lookupItemByBarcode items.LookupItemByBarcode) *LookupItemByBarcodeHandler {
	return &LookupItemByBarcodeHandler{lookupItemByBarcode: lookupItemByBarcode}
}

// Handle serves GET /api/v1/items/lookup?barcode=.
func (handler *LookupItemByBarcodeHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	result, err := handler.lookupItemByBarcode.Execute(request.Context(), items.LookupItemByBarcodeInput{
		Actor:   actorFrom(request),
		Barcode: request.URL.Query().Get("barcode"),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, dto.NewItemLookupResponse(result))
	return nil
}
