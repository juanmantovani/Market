package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/suppliers"
)

type DeleteSupplierHandler struct {
	deleteSupplier suppliers.DeleteSupplier
}

func NewDeleteSupplierHandler(deleteSupplier suppliers.DeleteSupplier) *DeleteSupplierHandler {
	return &DeleteSupplierHandler{deleteSupplier: deleteSupplier}
}

// Handle serves DELETE /api/v1/suppliers/{id}.
func (handler *DeleteSupplierHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	err := handler.deleteSupplier.Execute(request.Context(), suppliers.DeleteSupplierInput{
		Actor: actorFrom(request),
		ID:    request.PathValue("id"),
	})
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
