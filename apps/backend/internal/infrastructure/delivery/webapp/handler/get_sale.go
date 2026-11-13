package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/sales"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/sales"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type GetSaleHandler struct {
	getSale sales.GetSale
}

func NewGetSaleHandler(getSale sales.GetSale) *GetSaleHandler {
	return &GetSaleHandler{getSale: getSale}
}

// Handle serves GET /api/v1/sales/{id}.
func (handler *GetSaleHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	sale, err := handler.getSale.Execute(request.Context(), sales.GetSaleInput{
		Actor: actorFrom(request),
		ID:    request.PathValue("id"),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, dto.NewSaleResponse(sale))
	return nil
}
