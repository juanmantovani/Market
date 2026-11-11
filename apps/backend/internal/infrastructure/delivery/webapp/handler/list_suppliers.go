package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/suppliers"
	pagination "market/backend/internal/infrastructure/delivery/webapp/dto/pagination"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/suppliers"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type ListSuppliersHandler struct {
	listSuppliers suppliers.ListSuppliers
}

func NewListSuppliersHandler(listSuppliers suppliers.ListSuppliers) *ListSuppliersHandler {
	return &ListSuppliersHandler{listSuppliers: listSuppliers}
}

// Handle serves GET /api/v1/suppliers?q=&page=&limit=.
func (handler *ListSuppliersHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	page, limit, err := pageQuery(request)
	if err != nil {
		return err
	}

	query := request.URL.Query()
	result, err := handler.listSuppliers.Execute(request.Context(), suppliers.ListSuppliersInput{
		Actor:  actorFrom(request),
		Search: query.Get("q"),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, pagination.NewPageResponse(result, dto.NewSupplierResponse))
	return nil
}
