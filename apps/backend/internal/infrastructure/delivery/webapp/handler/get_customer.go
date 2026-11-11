package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/customers"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/customers"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type GetCustomerHandler struct {
	getCustomer customers.GetCustomer
}

func NewGetCustomerHandler(getCustomer customers.GetCustomer) *GetCustomerHandler {
	return &GetCustomerHandler{getCustomer: getCustomer}
}

// Handle serves GET /api/v1/customers/{id}.
func (handler *GetCustomerHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	customer, err := handler.getCustomer.Execute(request.Context(), customers.GetCustomerInput{
		Actor: actorFrom(request),
		ID:    request.PathValue("id"),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, dto.NewCustomerResponse(customer))
	return nil
}
