package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/sales"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/sales"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type VoidSaleHandler struct {
	voidSale sales.VoidSale
}

func NewVoidSaleHandler(voidSale sales.VoidSale) *VoidSaleHandler {
	return &VoidSaleHandler{voidSale: voidSale}
}

// Handle serves POST /api/v1/sales/{id}/void and answers the voided sale.
func (handler *VoidSaleHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	body, err := decodeJSON[dto.VoidSaleRequest](request)
	if err != nil {
		return err
	}

	sale, err := handler.voidSale.Execute(request.Context(), sales.VoidSaleInput{
		Actor:  actorFrom(request),
		ID:     request.PathValue("id"),
		Reason: body.Reason,
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, dto.NewSaleResponse(sale))
	return nil
}
