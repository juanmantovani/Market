package response

import (
	"errors"
	"log/slog"
	"net/http"

	"market/backend/internal/business/domain"
)

// WriteError is the only place that maps an error to an HTTP response. A
// *domain.Error is written with its own Code and Message and a status derived
// from its Kind; its Cause is logged, never exposed. Any other error is
// logged and hidden behind a generic 500.
func WriteError(w http.ResponseWriter, request *http.Request, logMsg string, err error) {
	var domainErr *domain.Error
	if errors.As(err, &domainErr) {
		if domainErr.Cause != nil {
			slog.ErrorContext(request.Context(), logMsg, "code", domainErr.Code, "error", domainErr.Cause)
		}
		WriteJSON(w, statusForKind(domainErr.Kind), ErrorResponse{
			Code: domainErr.Code, Message: domainErr.Message,
		})
		return
	}

	slog.ErrorContext(request.Context(), logMsg, "error", err)
	WriteJSON(w, http.StatusInternalServerError, ErrorResponse{
		Code: "internal_error", Message: "Ocurrió un error inesperado",
	})
}

func statusForKind(kind domain.Kind) int {
	switch kind {
	case domain.KindInvalid:
		return http.StatusBadRequest
	case domain.KindForbidden:
		return http.StatusForbidden
	case domain.KindConflict:
		return http.StatusConflict
	case domain.KindNotFound:
		return http.StatusNotFound
	case domain.KindUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
