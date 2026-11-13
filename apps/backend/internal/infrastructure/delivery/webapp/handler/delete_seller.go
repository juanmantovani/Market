package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/sellers"
)

type DeleteSellerHandler struct {
	deleteSeller sellers.DeleteSeller
}

func NewDeleteSellerHandler(deleteSeller sellers.DeleteSeller) *DeleteSellerHandler {
	return &DeleteSellerHandler{deleteSeller: deleteSeller}
}

// Handle serves DELETE /api/v1/sellers/{id}.
func (handler *DeleteSellerHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	err := handler.deleteSeller.Execute(request.Context(), sellers.DeleteSellerInput{
		Actor: actorFrom(request),
		ID:    request.PathValue("id"),
	})
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
