package dto

import "market/backend/internal/business/domain"

// PageResponse is the JSON shape of every paginated listing.
type PageResponse[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// NewPageResponse converts each item with convert. Items is never null in
// the JSON, so clients can iterate without a check.
func NewPageResponse[S, T any](page domain.Page[S], convert func(S) T) PageResponse[T] {
	items := make([]T, len(page.Items))
	for i, item := range page.Items {
		items[i] = convert(item)
	}

	return PageResponse[T]{Items: items, Total: page.Total, Page: page.Page, Limit: page.Limit}
}
