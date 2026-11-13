package handler

import (
	"net/http"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/sales"
	pagination "market/backend/internal/infrastructure/delivery/webapp/dto/pagination"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/sales"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type ListSalesHandler struct {
	listSales sales.ListSales
}

func NewListSalesHandler(listSales sales.ListSales) *ListSalesHandler {
	return &ListSalesHandler{listSales: listSales}
}

// Handle serves GET /api/v1/sales?from=&to=&customerId=&sellerId=&status=&page=&limit=.
// from and to are RFC 3339 instants (the client knows its time zone): from
// is included and to excluded.
func (handler *ListSalesHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	page, limit, err := pageQuery(request)
	if err != nil {
		return err
	}

	query := request.URL.Query()
	from, err := timeQuery(query.Get("from"))
	if err != nil {
		return err
	}
	to, err := timeQuery(query.Get("to"))
	if err != nil {
		return err
	}

	result, err := handler.listSales.Execute(request.Context(), sales.ListSalesInput{
		Actor: actorFrom(request),
		Filter: domain.SaleFilter{
			From:       from,
			To:         to,
			CustomerID: query.Get("customerId"),
			SellerID:   query.Get("sellerId"),
			Status:     domain.SaleStatus(query.Get("status")),
		},
		Page:  page,
		Limit: limit,
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, pagination.NewPageResponse(result, dto.NewSaleResponse))
	return nil
}

// timeQuery reads an optional RFC 3339 instant. Missing is the zero time,
// which doesn't filter.
func timeQuery(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	instant, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, domain.ErrSaleDateInvalid.WithCause(err)
	}

	return instant.UTC(), nil
}
