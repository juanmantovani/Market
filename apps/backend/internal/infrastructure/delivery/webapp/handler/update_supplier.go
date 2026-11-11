package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/suppliers"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/suppliers"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type UpdateSupplierHandler struct {
	updateSupplier suppliers.UpdateSupplier
}

func NewUpdateSupplierHandler(updateSupplier suppliers.UpdateSupplier) *UpdateSupplierHandler {
	return &UpdateSupplierHandler{updateSupplier: updateSupplier}
}

// Handle serves PUT /api/v1/suppliers/{id}: the body replaces every editable
// field, so an optional field sent as null is cleared.
func (handler *UpdateSupplierHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	body, err := decodeJSON[dto.SupplierRequest](request)
	if err != nil {
		return err
	}

	supplier, err := handler.updateSupplier.Execute(request.Context(), suppliers.UpdateSupplierInput{
		Actor:  actorFrom(request),
		ID:     request.PathValue("id"),
		Fields: body.Fields(),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, dto.NewSupplierResponse(supplier))
	return nil
}
