package clock

import (
	"testing"
	"time"
)

func TestSystemNowIsCurrentAndUTC(t *testing.T) {
	before := time.Now()
	now := System{}.Now()
	after := time.Now()

	if now.Location() != time.UTC {
		t.Fatalf("Now() location = %v, want UTC", now.Location())
	}
	if now.Before(before.Add(-time.Second)) || now.After(after.Add(time.Second)) {
		t.Fatalf("Now() = %v, want between %v and %v", now, before, after)
	}
}
