package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/sales"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/sales"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

// IdempotencyKeyHeader names the header a client sends so a retried request
// isn't processed twice.
const IdempotencyKeyHeader = "Idempotency-Key"

type CreateSaleHandler struct {
	createSale sales.CreateSale
}

func NewCreateSaleHandler(createSale sales.CreateSale) *CreateSaleHandler {
	return &CreateSaleHandler{createSale: createSale}
}

// Handle serves POST /api/v1/sales. A retry with the same Idempotency-Key
// and body answers the sale the first request created.
func (handler *CreateSaleHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	body, err := decodeJSON[dto.CreateSaleRequest](request)
	if err != nil {
		return err
	}

	sale, err := handler.createSale.Execute(request.Context(), sales.CreateSaleInput{
		Actor:   actorFrom(request),
		Request: body.Sale(request.Header.Get(IdempotencyKeyHeader)),
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusCreated, dto.NewSaleResponse(sale))
	return nil
}
