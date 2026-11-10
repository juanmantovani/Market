package gateway

import "time"

// Clock gives use cases the current time, so tests can fix it.
type Clock interface {
	Now() time.Time
}
