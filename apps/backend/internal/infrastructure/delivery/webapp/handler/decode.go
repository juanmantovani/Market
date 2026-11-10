package handler

import (
	"encoding/json"
	"net/http"

	"market/backend/internal/business/domain"
)

// decodeJSON decodes the request body into T, returning a *domain.Error
// (KindInvalid) when the body isn't valid JSON or exceeds MaxBodyBytes. The
// decoder's error travels as Cause so it reaches the log without being shown
// to the caller.
func decodeJSON[T any](request *http.Request) (T, error) {
	var body T
	if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
		return body, &domain.Error{
			Kind:    domain.KindInvalid,
			Code:    "invalid_body",
			Message: "Cuerpo de la solicitud inválido",
			Cause:   err,
		}
	}

	return body, nil
}
