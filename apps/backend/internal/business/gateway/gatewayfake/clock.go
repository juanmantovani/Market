package gatewayfake

import "time"

// Clock is a fake gateway.Clock that always returns Time.
type Clock struct {
	Time time.Time
}

func (clock Clock) Now() time.Time {
	return clock.Time
}
