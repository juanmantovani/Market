package handler

import (
	"net/http"
	"strconv"

	"market/backend/internal/business/domain"
	"market/backend/internal/infrastructure/delivery/webapp/middleware"
)

// pageQuery reads ?page= and ?limit=. A missing value is 0, which the use
// case turns into its default; anything that isn't a number is
// ErrInvalidPage.
func pageQuery(request *http.Request) (page, limit int, err error) {
	query := request.URL.Query()
	if page, err = intQuery(query.Get("page")); err != nil {
		return 0, 0, err
	}
	if limit, err = intQuery(query.Get("limit")); err != nil {
		return 0, 0, err
	}

	return page, limit, nil
}

func intQuery(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	number, err := strconv.Atoi(value)
	if err != nil {
		return 0, domain.ErrInvalidPage
	}

	return number, nil
}

// actorFrom returns the request's actor. Without one it returns the zero
// Actor, which every use case rejects as forbidden.
func actorFrom(request *http.Request) domain.Actor {
	actor, _ := middleware.ActorFromContext(request.Context())

	return actor
}
