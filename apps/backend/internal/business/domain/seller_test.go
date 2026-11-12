package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var sellerTime = time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)

func TestNewSeller(t *testing.T) {
	rate := ptr[Percent](250)
	seller, err := NewSeller(SellerFields{Name: "  Juan Pérez ", Code: " jp ", CommissionRate: rate}, sellerTime)
	if err != nil {
		t.Fatalf("NewSeller() error = %v", err)
	}

	if seller.Name != "Juan Pérez" || seller.Code != "JP" || seller.CommissionRate == nil || *seller.CommissionRate != 250 {
		t.Errorf("seller = %+v", seller.SellerFields)
	}
	if seller.CommissionRate == rate {
		t.Error("the seller keeps the caller's commission pointer")
	}
	if !seller.CreatedAt.Equal(sellerTime) || !seller.UpdatedAt.Equal(sellerTime) || seller.ID != "" {
		t.Errorf("seller = %+v", seller)
	}
}

func TestNewSellerAcceptsTheEdges(t *testing.T) {
	for _, fields := range []SellerFields{
		{Name: "Ana", Code: "A"},
		{Name: "Beto", Code: "V12345", CommissionRate: ptr[Percent](0)},
		{Name: "Carla", Code: "c1", CommissionRate: ptr[Percent](PercentScale)},
	} {
		if _, err := NewSeller(fields, sellerTime); err != nil {
			t.Errorf("NewSeller(%+v) error = %v", fields, err)
		}
	}
}

func TestNewSellerValidates(t *testing.T) {
	tests := []struct {
		name   string
		fields SellerFields
		want   *Error
	}{
		{"missing name", SellerFields{Name: " ", Code: "JP"}, ErrSellerNameRequired},
		{"long name", SellerFields{Name: strings.Repeat("n", MaxSellerNameLength+1), Code: "JP"}, ErrSellerNameTooLong},
		{"missing code", SellerFields{Name: "Juan", Code: "  "}, ErrSellerCodeRequired},
		{"long code", SellerFields{Name: "Juan", Code: "ABCDEFG"}, ErrSellerCodeInvalid},
		{"code with a space", SellerFields{Name: "Juan", Code: "J P"}, ErrSellerCodeInvalid},
		{"code with a symbol", SellerFields{Name: "Juan", Code: "J-P"}, ErrSellerCodeInvalid},
		{"code with an accent", SellerFields{Name: "Juan", Code: "JOSÉ"}, ErrSellerCodeInvalid},
		{"negative commission", SellerFields{Name: "Juan", Code: "JP", CommissionRate: ptr[Percent](-1)}, ErrSellerCommissionInvalid},
		{"commission above 100 %", SellerFields{Name: "Juan", Code: "JP", CommissionRate: ptr[Percent](PercentScale + 1)}, ErrSellerCommissionInvalid},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewSeller(test.fields, sellerTime); !errors.Is(err, test.want) {
				t.Fatalf("NewSeller() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestSellerReplace(t *testing.T) {
	seller, _ := NewSeller(SellerFields{Name: "Juan", Code: "JP", CommissionRate: ptr[Percent](250)}, sellerTime)
	seller.ID = "s1"
	later := sellerTime.Add(time.Hour)

	replaced, err := seller.Replace(SellerFields{Name: "Juan Pablo", Code: "jpa"}, later)
	if err != nil {
		t.Fatalf("Replace() error = %v", err)
	}
	if replaced.ID != "s1" || replaced.Name != "Juan Pablo" || replaced.Code != "JPA" || replaced.CommissionRate != nil ||
		!replaced.CreatedAt.Equal(sellerTime) || !replaced.UpdatedAt.Equal(later) {
		t.Fatalf("replaced = %+v", replaced)
	}

	if _, err := seller.Replace(SellerFields{Name: "Juan"}, later); !errors.Is(err, ErrSellerCodeRequired) {
		t.Fatalf("Replace(without code) error = %v", err)
	}
}
