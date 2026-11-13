package handler

import (
	"net/http"

	dto "market/backend/internal/infrastructure/delivery/webapp/dto/health"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

// Live reports that the process is up and serving requests.
func Live(w http.ResponseWriter, _ *http.Request) {
	response.WriteJSON(w, http.StatusOK, dto.StatusResponse{Status: "ok"})
}

// NotFound answers API paths no route matched with the same JSON error shape
// as every other failure, instead of the mux's plain-text 404.
func NotFound(w http.ResponseWriter, _ *http.Request) {
	response.WriteJSON(w, http.StatusNotFound, response.ErrorResponse{
		Code: "route_not_found", Message: "La ruta solicitada no existe",
	})
}
