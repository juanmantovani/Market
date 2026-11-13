package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/sellers"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/sellers"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type CreateSellerHandler struct {
	createSeller sellers.CreateSeller
}

func NewCreateSellerHandler(createSeller sellers.CreateSeller) *CreateSellerHandler {
	return &CreateSellerHandler{createSeller: createSeller}
}

// Handle serves POST /api/v1/sellers.
func (handler *CreateSellerHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	body, err := decodeJSON[dto.SellerRequest](request)
	if err != nil {
		return err
	}

	seller, err := handler.createSeller.Execute(request.Context(), sellers.CreateSellerInput{
		Actor:  actorFrom(request),
		Fields: body.Fields(),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusCreated, dto.NewSellerResponse(seller))
	return nil
}
