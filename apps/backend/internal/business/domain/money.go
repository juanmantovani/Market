package domain

import "math/big"

var ErrAmountOutOfRange = &Error{Kind: KindInvalid, Code: "amount_out_of_range", Message: "El importe excede el máximo permitido"}

// Money is an amount in cents of the store currency (ARS). Money is never
// represented as floating point, in Go, JSON or TypeScript.
type Money int64

// Percent is expressed in hundredths of a percent: 12.5 % is 1250.
type Percent int64

const PercentScale = 10_000

// IsValid reports whether percent lies between 0 % and 100 %.
func (percent Percent) IsValid() bool {
	return percent >= 0 && percent <= PercentScale
}

// Times returns the amount for quantity at this unit price, rounded half
// away from zero to the cent.
func (price Money) Times(quantity Quantity) (Money, error) {
	return mulDivRound(int64(price), int64(quantity), QuantityScale)
}

// Percent returns percent of the amount, rounded half away from zero to the
// cent.
func (amount Money) Percent(percent Percent) (Money, error) {
	return mulDivRound(int64(amount), int64(percent), PercentScale)
}

// mulDivRound computes a*b/divisor without intermediate overflow. divisor
// must be positive.
func mulDivRound(a, b, divisor int64) (Money, error) {
	product := new(big.Int).Mul(big.NewInt(a), big.NewInt(b))
	quotient, remainder := new(big.Int).QuoRem(product, big.NewInt(divisor), new(big.Int))

	doubledRemainder := new(big.Int).Lsh(new(big.Int).Abs(remainder), 1)
	if doubledRemainder.Cmp(big.NewInt(divisor)) >= 0 {
		if product.Sign() < 0 {
			quotient.Sub(quotient, big.NewInt(1))
		} else {
			quotient.Add(quotient, big.NewInt(1))
		}
	}

	if !quotient.IsInt64() {
		return 0, ErrAmountOutOfRange
	}

	return Money(quotient.Int64()), nil
}
