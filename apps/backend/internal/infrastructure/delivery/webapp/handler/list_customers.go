package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/customers"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/customers"
	pagination "market/backend/internal/infrastructure/delivery/webapp/dto/pagination"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type ListCustomersHandler struct {
	listCustomers customers.ListCustomers
}

func NewListCustomersHandler(listCustomers customers.ListCustomers) *ListCustomersHandler {
	return &ListCustomersHandler{listCustomers: listCustomers}
}

// Handle serves GET /api/v1/customers?q=&page=&limit=.
func (handler *ListCustomersHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	page, limit, err := pageQuery(request)
	if err != nil {
		return err
	}

	query := request.URL.Query()
	result, err := handler.listCustomers.Execute(request.Context(), customers.ListCustomersInput{
		Actor:  actorFrom(request),
		Search: query.Get("q"),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, pagination.NewPageResponse(result, dto.NewCustomerResponse))
	return nil
}
