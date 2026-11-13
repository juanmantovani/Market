package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/customers"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/customers"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type UpdateCustomerHandler struct {
	updateCustomer customers.UpdateCustomer
}

func NewUpdateCustomerHandler(updateCustomer customers.UpdateCustomer) *UpdateCustomerHandler {
	return &UpdateCustomerHandler{updateCustomer: updateCustomer}
}

// Handle serves PUT /api/v1/customers/{id}: the body replaces every editable
// field, so an optional field sent as null is cleared.
func (handler *UpdateCustomerHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	body, err := decodeJSON[dto.CustomerRequest](request)
	if err != nil {
		return err
	}

	customer, err := handler.updateCustomer.Execute(request.Context(), customers.UpdateCustomerInput{
		Actor:  actorFrom(request),
		ID:     request.PathValue("id"),
		Fields: body.Fields(),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, dto.NewCustomerResponse(customer))
	return nil
}
