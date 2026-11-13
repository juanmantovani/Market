package domain

import (
	"errors"
	"math"
	"testing"
)

func TestMoneyTimes(t *testing.T) {
	tests := []struct {
		name     string
		price    Money
		quantity Quantity
		want     Money
	}{
		{name: "whole units", price: 1200, quantity: WholeQuantity(2), want: 2400},
		{name: "fraction by weight", price: 250_000, quantity: 750, want: 187_500},
		{name: "rounds up from half a cent", price: 100, quantity: 5, want: 1},
		{name: "rounds down below half a cent", price: 100, quantity: 4, want: 0},
		{name: "rounds to the nearest cent", price: 999, quantity: 333, want: 333},
		{name: "negative rounds away from zero", price: -100, quantity: 5, want: -1},
		{name: "zero quantity", price: 1500, quantity: 0, want: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.price.Times(test.quantity)
			if err != nil {
				t.Fatalf("Times() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("%d.Times(%d) = %d, want %d", test.price, test.quantity, got, test.want)
			}
		})
	}
}

func TestMoneyTimesRejectsOverflow(t *testing.T) {
	_, err := Money(math.MaxInt64).Times(WholeQuantity(2))
	if !errors.Is(err, ErrAmountOutOfRange) {
		t.Fatalf("Times() error = %v, want ErrAmountOutOfRange", err)
	}
}

func TestMoneyPercent(t *testing.T) {
	tests := []struct {
		amount  Money
		percent Percent
		want    Money
	}{
		{amount: 10_000, percent: 1250, want: 1250},
		{amount: 333, percent: 1000, want: 33},
		{amount: 335, percent: 1000, want: 34},
		{amount: 4275, percent: PercentScale, want: 4275},
		{amount: 4275, percent: 0, want: 0},
	}

	for _, test := range tests {
		got, err := test.amount.Percent(test.percent)
		if err != nil {
			t.Fatalf("Percent() error = %v", err)
		}
		if got != test.want {
			t.Errorf("%d.Percent(%d) = %d, want %d", test.amount, test.percent, got, test.want)
		}
	}
}

func TestPercentIsValid(t *testing.T) {
	for _, percent := range []Percent{0, 1, 1250, PercentScale} {
		if !percent.IsValid() {
			t.Errorf("Percent(%d).IsValid() = false, want true", percent)
		}
	}
	for _, percent := range []Percent{-1, PercentScale + 1} {
		if percent.IsValid() {
			t.Errorf("Percent(%d).IsValid() = true, want false", percent)
		}
	}
}
