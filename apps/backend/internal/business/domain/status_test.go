package domain

import (
	"errors"
	"testing"
)

func TestParseStatusFilter(t *testing.T) {
	tests := map[string]StatusFilter{
		"":         StatusActive,
		"active":   StatusActive,
		"inactive": StatusInactive,
		"all":      StatusAll,
	}
	for input, want := range tests {
		got, err := ParseStatusFilter(input)
		if err != nil || got != want {
			t.Errorf("ParseStatusFilter(%q) = %q, %v; want %q", input, got, err, want)
		}
	}

	if _, err := ParseStatusFilter("deleted"); !errors.Is(err, ErrInvalidStatusFilter) {
		t.Fatalf("ParseStatusFilter(deleted) error = %v, want ErrInvalidStatusFilter", err)
	}
}

func TestStatusFilterMatches(t *testing.T) {
	tests := []struct {
		filter StatusFilter
		active bool
		want   bool
	}{
		{StatusActive, true, true},
		{StatusActive, false, false},
		{StatusInactive, true, false},
		{StatusInactive, false, true},
		{StatusAll, true, true},
		{StatusAll, false, true},
	}
	for _, test := range tests {
		if got := test.filter.Matches(test.active); got != test.want {
			t.Errorf("%q.Matches(%v) = %v, want %v", test.filter, test.active, got, test.want)
		}
	}
}
