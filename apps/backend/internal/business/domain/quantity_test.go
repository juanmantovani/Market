package domain

import "testing"

func TestWholeQuantity(t *testing.T) {
	if got := WholeQuantity(3); got != 3000 {
		t.Fatalf("WholeQuantity(3) = %d, want 3000", got)
	}
	if !WholeQuantity(7).IsWhole() {
		t.Fatal("WholeQuantity(7).IsWhole() = false")
	}
	if Quantity(750).IsWhole() {
		t.Fatal("Quantity(750).IsWhole() = true")
	}
}

func TestUnitIsValid(t *testing.T) {
	for _, unit := range []Unit{UnitPiece, UnitKilogram, UnitLiter, UnitMeter} {
		if !unit.IsValid() {
			t.Errorf("%q.IsValid() = false, want true", unit)
		}
	}
	for _, unit := range []Unit{"", "KG", "box"} {
		if unit.IsValid() {
			t.Errorf("%q.IsValid() = true, want false", unit)
		}
	}
}

func TestUnitAccepts(t *testing.T) {
	tests := []struct {
		unit     Unit
		quantity Quantity
		want     bool
	}{
		{unit: UnitPiece, quantity: WholeQuantity(2), want: true},
		{unit: UnitPiece, quantity: 1500, want: false},
		{unit: UnitKilogram, quantity: 250, want: true},
		{unit: UnitLiter, quantity: 1, want: true},
		{unit: UnitMeter, quantity: 0, want: false},
		{unit: UnitKilogram, quantity: -250, want: false},
		{unit: "box", quantity: WholeQuantity(1), want: false},
	}

	for _, test := range tests {
		if got := test.unit.Accepts(test.quantity); got != test.want {
			t.Errorf("%q.Accepts(%d) = %v, want %v", test.unit, test.quantity, got, test.want)
		}
	}
}

func TestUnitAllowsFraction(t *testing.T) {
	if UnitPiece.AllowsFraction() {
		t.Error("UnitPiece.AllowsFraction() = true")
	}
	if !UnitKilogram.AllowsFraction() {
		t.Error("UnitKilogram.AllowsFraction() = false")
	}
	if Unit("box").AllowsFraction() {
		t.Error(`Unit("box").AllowsFraction() = true`)
	}
}
