package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/sellers"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/sellers"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type GetSellerHandler struct {
	getSeller sellers.GetSeller
}

func NewGetSellerHandler(getSeller sellers.GetSeller) *GetSellerHandler {
	return &GetSellerHandler{getSeller: getSeller}
}

// Handle serves GET /api/v1/sellers/{id}.
func (handler *GetSellerHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	seller, err := handler.getSeller.Execute(request.Context(), sellers.GetSellerInput{
		Actor: actorFrom(request),
		ID:    request.PathValue("id"),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, dto.NewSellerResponse(seller))
	return nil
}
