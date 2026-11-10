package domain

import (
	"errors"
	"testing"
)

func TestNewPageRequest(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		limit      int
		want       PageRequest
		wantOffset int
	}{
		{name: "defaults", want: PageRequest{Page: 1, Limit: DefaultPageLimit}, wantOffset: 0},
		{name: "explicit", page: 3, limit: 10, want: PageRequest{Page: 3, Limit: 10}, wantOffset: 20},
		{name: "max limit", page: 1, limit: MaxPageLimit, want: PageRequest{Page: 1, Limit: MaxPageLimit}, wantOffset: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NewPageRequest(test.page, test.limit)
			if err != nil {
				t.Fatalf("NewPageRequest() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("NewPageRequest() = %+v, want %+v", got, test.want)
			}
			if got.Offset() != test.wantOffset {
				t.Fatalf("Offset() = %d, want %d", got.Offset(), test.wantOffset)
			}
		})
	}
}

func TestNewPageRequestRejectsOutOfRange(t *testing.T) {
	for _, input := range [][2]int{{-1, 10}, {1, -5}, {1, MaxPageLimit + 1}} {
		if _, err := NewPageRequest(input[0], input[1]); !errors.Is(err, ErrInvalidPage) {
			t.Errorf("NewPageRequest(%d, %d) error = %v, want ErrInvalidPage", input[0], input[1], err)
		}
	}
}
