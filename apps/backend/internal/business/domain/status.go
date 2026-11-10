package domain

// StatusFilter selects records of a listing by whether they are active.
// Catalog and third-party records are never deleted, only deactivated.
type StatusFilter string

const (
	StatusActive   StatusFilter = "active"
	StatusInactive StatusFilter = "inactive"
	StatusAll      StatusFilter = "all"
)

var ErrInvalidStatusFilter = &Error{Kind: KindInvalid, Code: "invalid_status_filter", Message: "El filtro de estado no es válido"}

// ParseStatusFilter reads a status filter, defaulting to active records.
func ParseStatusFilter(value string) (StatusFilter, error) {
	switch filter := StatusFilter(value); filter {
	case "":
		return StatusActive, nil
	case StatusActive, StatusInactive, StatusAll:
		return filter, nil
	default:
		return "", ErrInvalidStatusFilter
	}
}

func (filter StatusFilter) Matches(active bool) bool {
	switch filter {
	case StatusActive:
		return active
	case StatusInactive:
		return !active
	default:
		return true
	}
}
