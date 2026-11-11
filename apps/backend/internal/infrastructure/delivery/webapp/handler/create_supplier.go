package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/suppliers"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/suppliers"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type CreateSupplierHandler struct {
	createSupplier suppliers.CreateSupplier
}

func NewCreateSupplierHandler(createSupplier suppliers.CreateSupplier) *CreateSupplierHandler {
	return &CreateSupplierHandler{createSupplier: createSupplier}
}

// Handle serves POST /api/v1/suppliers.
func (handler *CreateSupplierHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	body, err := decodeJSON[dto.SupplierRequest](request)
	if err != nil {
		return err
	}

	supplier, err := handler.createSupplier.Execute(request.Context(), suppliers.CreateSupplierInput{
		Actor:  actorFrom(request),
		Fields: body.Fields(),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusCreated, dto.NewSupplierResponse(supplier))
	return nil
}
