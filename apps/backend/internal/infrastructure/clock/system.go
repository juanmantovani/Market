package clock

import (
	"time"

	"market/backend/internal/business/gateway"
)

var _ gateway.Clock = System{}

// System is the wall clock, in UTC so stored timestamps don't depend on the
// server's time zone.
type System struct{}

func (System) Now() time.Time {
	return time.Now().UTC()
}
