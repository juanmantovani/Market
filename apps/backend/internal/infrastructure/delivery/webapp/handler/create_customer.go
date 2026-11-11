package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/customers"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/customers"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type CreateCustomerHandler struct {
	createCustomer customers.CreateCustomer
}

func NewCreateCustomerHandler(createCustomer customers.CreateCustomer) *CreateCustomerHandler {
	return &CreateCustomerHandler{createCustomer: createCustomer}
}

// Handle serves POST /api/v1/customers.
func (handler *CreateCustomerHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	body, err := decodeJSON[dto.CustomerRequest](request)
	if err != nil {
		return err
	}

	customer, err := handler.createCustomer.Execute(request.Context(), customers.CreateCustomerInput{
		Actor:  actorFrom(request),
		Fields: body.Fields(),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusCreated, dto.NewCustomerResponse(customer))
	return nil
}
