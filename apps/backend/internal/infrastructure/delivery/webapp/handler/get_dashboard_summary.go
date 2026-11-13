package handler

import (
	"net/http"

	"market/backend/internal/business/usecase/dashboard"
	dto "market/backend/internal/infrastructure/delivery/webapp/dto/dashboard"
	"market/backend/internal/infrastructure/delivery/webapp/response"
)

type GetDashboardSummaryHandler struct {
	getSummary dashboard.GetSummary
}

func NewGetDashboardSummaryHandler(getSummary dashboard.GetSummary) *GetDashboardSummaryHandler {
	return &GetDashboardSummaryHandler{getSummary: getSummary}
}

// Handle serves GET /api/v1/dashboard/summary?from=&to=. from and to are
// required RFC 3339 instants that bound the period (from included, to
// excluded): the client sends the start and end of its own day, so the
// summary follows its time zone.
func (handler *GetDashboardSummaryHandler) Handle(w http.ResponseWriter, request *http.Request) error {
	query := request.URL.Query()
	from, err := timeQuery(query.Get("from"))
	if err != nil {
		return err
	}
	to, err := timeQuery(query.Get("to"))
	if err != nil {
		return err
	}

	summary, err := handler.getSummary.Execute(request.Context(), dashboard.GetSummaryInput{
		Actor: actorFrom(request),
		From:  from,
		To:    to,
	})
	if err != nil {
		return err
	}

	response.WriteJSON(w, http.StatusOK, dto.NewSummaryResponse(summary))
	return nil
}
