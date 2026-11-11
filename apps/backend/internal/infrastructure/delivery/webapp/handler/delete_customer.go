package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/customers"
)

type DeleteCustomerHandler struct {
	deleteCustomer customers.DeleteCustomer
}

func NewDeleteCustomerHandler(deleteCustomer customers.DeleteCustomer) *DeleteCustomerHandler {
	return &DeleteCustomerHandler{deleteCustomer: deleteCustomer}
}

// Handle serves DELETE /api/v1/customers/{id}.
func (handler *DeleteCustomerHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	err := handler.deleteCustomer.Execute(request.Context(), customers.DeleteCustomerInput{
		Actor: actorFrom(request),
		ID:    request.PathValue("id"),
	})
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
