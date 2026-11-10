package domain

const (
	DefaultPageLimit = 25
	MaxPageLimit     = 100
)

var ErrInvalidPage = &Error{Kind: KindInvalid, Code: "invalid_page", Message: "La paginación solicitada no es válida"}

// PageRequest asks for one page of a listing. Pages start at 1.
type PageRequest struct {
	Page  int
	Limit int
}

// NewPageRequest applies the defaults for a page or limit left at zero and
// rejects anything outside 1..MaxPageLimit.
func NewPageRequest(page, limit int) (PageRequest, error) {
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = DefaultPageLimit
	}
	if page < 1 || limit < 1 || limit > MaxPageLimit {
		return PageRequest{}, ErrInvalidPage
	}

	return PageRequest{Page: page, Limit: limit}, nil
}

func (request PageRequest) Offset() int {
	return (request.Page - 1) * request.Limit
}

// Page is one page of a listing plus the total across all pages.
type Page[T any] struct {
	Items []T
	Total int
	Page  int
	Limit int
}
