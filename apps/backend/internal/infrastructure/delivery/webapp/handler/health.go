package handler

import (
	"context"
	"net/http"

	"market/backend/internal/business/domain"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/health"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

// Live reports that the process is up and serving requests.
func Live(w http.ResponseWriter, _ *http.Request) {
	response.WriteJSON(w, http.StatusOK, dto.StatusResponse{Status: "ok"})
}

// ReadyHandler reports whether the storage behind the API is reachable.
type ReadyHandler struct {
	ping func(ctx context.Context) error
}

func NewReadyHandler(ping func(ctx context.Context) error) *ReadyHandler {
	return &ReadyHandler{ping: ping}
}

func (handler *ReadyHandler) Handle(w http.ResponseWriter, r *http.Request) error {
	if err := handler.ping(r.Context()); err != nil {
		return domain.ErrStorageUnavailable.WithCause(err)
	}

	response.WriteJSON(w, http.StatusOK, dto.StatusResponse{Status: "ok"})
	return nil
}

// NotFound answers API paths no route matched with the same JSON error shape
// as every other failure, instead of the mux's plain-text 404.
func NotFound(w http.ResponseWriter, _ *http.Request) {
	response.WriteJSON(w, http.StatusNotFound, response.ErrorResponse{
		Code: "route_not_found", Message: "La ruta solicitada no existe",
	})
}
