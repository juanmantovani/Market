package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/sellers"
	pagination "market/backend/internal/infrastructure/delivery/webapp/dto/pagination"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/sellers"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type ListSellersHandler struct {
	listSellers sellers.ListSellers
}

func NewListSellersHandler(listSellers sellers.ListSellers) *ListSellersHandler {
	return &ListSellersHandler{listSellers: listSellers}
}

// Handle serves GET /api/v1/sellers?q=&page=&limit=.
func (handler *ListSellersHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	page, limit, err := pageQuery(request)
	if err != nil {
		return err
	}

	query := request.URL.Query()
	result, err := handler.listSellers.Execute(request.Context(), sellers.ListSellersInput{
		Actor:  actorFrom(request),
		Search: query.Get("q"),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, pagination.NewPageResponse(result, dto.NewSellerResponse))
	return nil
}
