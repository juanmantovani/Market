package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/sellers"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/sellers"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type UpdateSellerHandler struct {
	updateSeller sellers.UpdateSeller
}

func NewUpdateSellerHandler(updateSeller sellers.UpdateSeller) *UpdateSellerHandler {
	return &UpdateSellerHandler{updateSeller: updateSeller}
}

// Handle serves PUT /api/v1/sellers/{id}: the body replaces every editable
// field, so an optional field sent as null is cleared.
func (handler *UpdateSellerHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	body, err := decodeJSON[dto.SellerRequest](request)
	if err != nil {
		return err
	}

	seller, err := handler.updateSeller.Execute(request.Context(), sellers.UpdateSellerInput{
		Actor:  actorFrom(request),
		ID:     request.PathValue("id"),
		Fields: body.Fields(),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, dto.NewSellerResponse(seller))
	return nil
}
