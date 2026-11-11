package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/suppliers"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/suppliers"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type GetSupplierHandler struct {
	getSupplier suppliers.GetSupplier
}

func NewGetSupplierHandler(getSupplier suppliers.GetSupplier) *GetSupplierHandler {
	return &GetSupplierHandler{getSupplier: getSupplier}
}

// Handle serves GET /api/v1/suppliers/{id}.
func (handler *GetSupplierHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	supplier, err := handler.getSupplier.Execute(request.Context(), suppliers.GetSupplierInput{
		Actor: actorFrom(request),
		ID:    request.PathValue("id"),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, dto.NewSupplierResponse(supplier))
	return nil
}
