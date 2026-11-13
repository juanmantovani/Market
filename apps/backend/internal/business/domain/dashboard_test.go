package domain

import (
	"testing"
	"time"
)

func TestNewSummaryPeriod(t *testing.T) {
	from := time.Date(2026, 11, 12, 3, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		from time.Time
		to   time.Time
		ok   bool
	}{
		{name: "a day", from: from, to: from.Add(24 * time.Hour), ok: true},
		{name: "the longest period", from: from, to: from.Add(MaxSummaryPeriod), ok: true},
		{name: "too long", from: from, to: from.Add(MaxSummaryPeriod + time.Second)},
		{name: "empty", from: from, to: from},
		{name: "reversed", from: from.Add(time.Hour), to: from},
		{name: "no start", to: from},
		{name: "no end", from: from},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			period, err := NewSummaryPeriod(test.from, test.to)

			if test.ok {
				if err != nil || period.From != test.from || period.To != test.to {
					t.Fatalf("NewSummaryPeriod() = %+v, %v", period, err)
				}
				return
			}
			if err != ErrSummaryPeriodInvalid {
				t.Fatalf("error = %v, want %s", err, ErrSummaryPeriodInvalid.Code)
			}
		})
	}
}

func TestNewSalesSummary(t *testing.T) {
	sales := []Sale{
		{Status: SaleCompleted, Total: 100_000, Change: 50_000, Payments: []Payment{{Method: PaymentCash, Amount: 150_000}}},
		{Status: SaleCompleted, Total: 200_000, Payments: []Payment{{Method: PaymentQR, Amount: 120_000}, {Method: PaymentCash, Amount: 80_000}}},
		{Status: SaleCompleted, Total: 30_000, Payments: []Payment{{Method: PaymentDebitCard, Amount: 30_000}}},
		{Status: SaleVoided, Total: 999_900, Payments: []Payment{{Method: PaymentTransfer, Amount: 999_900}}},
	}

	got := NewSalesSummary(sales)

	if got.Count != 3 || got.Total != 330_000 {
		t.Fatalf("count, total = %d, %d", got.Count, got.Total)
	}
	want := []MethodTotal{
		{Method: PaymentCash, Amount: 180_000},
		{Method: PaymentDebitCard, Amount: 30_000},
		{Method: PaymentQR, Amount: 120_000},
	}
	if len(got.ByMethod) != len(want) {
		t.Fatalf("ByMethod = %+v, want %+v", got.ByMethod, want)
	}
	var sum Money
	for i, total := range got.ByMethod {
		if total != want[i] {
			t.Errorf("ByMethod[%d] = %+v, want %+v", i, total, want[i])
		}
		sum += total.Amount
	}
	if sum != got.Total {
		t.Errorf("methods add up to %d, want %d", sum, got.Total)
	}
}

func TestNewSalesSummaryWithoutSales(t *testing.T) {
	got := NewSalesSummary(nil)

	if got.Count != 0 || got.Total != 0 || len(got.ByMethod) != 0 {
		t.Fatalf("summary = %+v", got)
	}
}
