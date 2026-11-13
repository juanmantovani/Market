package domain

import (
	"fmt"
	"math/big"
	"slices"
)

// DiscountType says how a discount's Value reads.
type DiscountType string

const (
	// DiscountPercent takes a percent off: Value is in hundredths of a
	// percent, like Percent (10 % is 1000).
	DiscountPercent DiscountType = "percent"
	// DiscountAmount takes a fixed amount off: Value is in cents.
	DiscountAmount DiscountType = "amount"
)

var (
	ErrSaleDiscountInvalid      = &Error{Kind: KindInvalid, Code: "sale_discount_invalid", Message: "El descuento tiene que ser un porcentaje de 0 a 100 % o un monto que no sea negativo"}
	ErrSaleDiscountExceedsLine  = &Error{Kind: KindInvalid, Code: "sale_discount_exceeds_line", Message: "El descuento de una línea no puede superar su importe"}
	ErrSaleDiscountExceedsTotal = &Error{Kind: KindInvalid, Code: "sale_discount_exceeds_total", Message: "El descuento general no puede superar el total de la venta"}
)

// Discount is taken off a sale line or off the whole sale (the cart).
type Discount struct {
	Type  DiscountType
	Value int64
}

// validateDiscount checks an optional discount and returns a copy that
// doesn't alias the caller's; nil means no discount.
func validateDiscount(discount *Discount) (*Discount, error) {
	if discount == nil {
		return nil, nil
	}
	switch discount.Type {
	case DiscountPercent:
		if !Percent(discount.Value).IsValid() {
			return nil, ErrSaleDiscountInvalid
		}
	case DiscountAmount:
		if discount.Value < 0 || discount.Value > int64(MaxSaleAmount) {
			return nil, ErrSaleDiscountInvalid
		}
	default:
		return nil, ErrSaleDiscountInvalid
	}
	valid := *discount

	return &valid, nil
}

// amountOff is what discount takes off amount: the percent rounded to the
// cent, or the fixed amount, which can't be more than amount (exceeds is the
// error then). A nil discount takes nothing.
func (discount *Discount) amountOff(amount Money, exceeds *Error) (Money, error) {
	if discount == nil {
		return 0, nil
	}
	if discount.Type == DiscountPercent {
		return amount.Percent(Percent(discount.Value))
	}
	if Money(discount.Value) > amount {
		return 0, exceeds
	}

	return Money(discount.Value), nil
}

// String identifies the discount in a request fingerprint.
func (discount *Discount) String() string {
	if discount == nil {
		return "none"
	}

	return fmt.Sprintf("%s:%d", discount.Type, discount.Value)
}

// prorate splits total among weights in proportion to each one, by largest
// remainder: every share is its exact share rounded down, and the cents left
// go one each to the largest fractions, the first line winning a tie. The
// shares add up to total, and while total is at most the sum of the weights
// no share is larger than its weight, so no line ends up negative.
func prorate(total Money, weights []Money) []Money {
	shares := make([]Money, len(weights))
	var sum Money
	for _, weight := range weights {
		sum += weight
	}
	if total == 0 || sum == 0 {
		return shares
	}

	remainders := make([]*big.Int, len(weights))
	left := total
	for i, weight := range weights {
		share, remainder := new(big.Int).QuoRem(
			new(big.Int).Mul(big.NewInt(int64(total)), big.NewInt(int64(weight))),
			big.NewInt(int64(sum)),
			new(big.Int),
		)
		shares[i] = Money(share.Int64())
		remainders[i] = remainder
		left -= shares[i]
	}

	order := make([]int, len(weights))
	for i := range order {
		order[i] = i
	}
	slices.SortStableFunc(order, func(a, b int) int {
		return remainders[b].Cmp(remainders[a])
	})
	for _, i := range order[:int(left)] {
		shares[i]++
	}

	return shares
}
