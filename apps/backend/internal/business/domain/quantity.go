package domain

// Quantity is measured in thousandths of a unit, so 0.250 kg is 250 and
// three pieces are 3000.
type Quantity int64

const QuantityScale = 1000

func WholeQuantity(units int64) Quantity {
	return Quantity(units * QuantityScale)
}

func (quantity Quantity) IsWhole() bool {
	return quantity%QuantityScale == 0
}

// Unit is how an item is measured when it is sold.
type Unit string

const (
	UnitPiece    Unit = "unit"
	UnitKilogram Unit = "kg"
	UnitLiter    Unit = "l"
	UnitMeter    Unit = "m"
)

func (unit Unit) IsValid() bool {
	switch unit {
	case UnitPiece, UnitKilogram, UnitLiter, UnitMeter:
		return true
	default:
		return false
	}
}

func (unit Unit) AllowsFraction() bool {
	return unit.IsValid() && unit != UnitPiece
}

// Accepts reports whether quantity can be sold in this unit: always
// positive, and whole for items sold by the piece.
func (unit Unit) Accepts(quantity Quantity) bool {
	if !unit.IsValid() || quantity <= 0 {
		return false
	}

	return unit.AllowsFraction() || quantity.IsWhole()
}
