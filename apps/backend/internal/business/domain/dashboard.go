package domain

import "time"

// MaxSummaryPeriod caps the period of a sales summary, which is read by
// paging through every sale in it.
const MaxSummaryPeriod = 31 * 24 * time.Hour

var ErrSummaryPeriodInvalid = &Error{Kind: KindInvalid, Code: "summary_period_invalid", Message: "El período del resumen no es válido"}

// paymentMethodOrder is the order the summary lists the payment methods in.
var paymentMethodOrder = []PaymentMethod{PaymentCash, PaymentDebitCard, PaymentCreditCard, PaymentTransfer, PaymentQR}

// SummaryPeriod is the period a summary covers: From included, To excluded.
// The client picks it, because "today" depends on its time zone.
type SummaryPeriod struct {
	From time.Time
	To   time.Time
}

// NewSummaryPeriod requires both bounds, in order and at most
// MaxSummaryPeriod apart.
func NewSummaryPeriod(from, to time.Time) (SummaryPeriod, error) {
	if from.IsZero() || to.IsZero() || !from.Before(to) || to.Sub(from) > MaxSummaryPeriod {
		return SummaryPeriod{}, ErrSummaryPeriodInvalid
	}

	return SummaryPeriod{From: from, To: to}, nil
}

// MethodTotal is what one payment method collected.
type MethodTotal struct {
	Method PaymentMethod
	Amount Money
}

// SalesSummary adds up the completed sales of a period. Voided sales are
// left out by whoever feeds it.
type SalesSummary struct {
	Count int
	Total Money
	// ByMethod lists, in a fixed order, the methods that collected something.
	// Cash is net of the change given back, so the methods add up to Total.
	ByMethod []MethodTotal
}

// NewSalesSummary adds up sales, ignoring any that is voided.
func NewSalesSummary(sales []Sale) SalesSummary {
	var summary SalesSummary
	totals := make(map[PaymentMethod]Money, len(paymentMethodOrder))
	for _, sale := range sales {
		if sale.Status != SaleCompleted {
			continue
		}
		summary.Count++
		summary.Total += sale.Total
		for _, payment := range sale.Payments {
			totals[payment.Method] += payment.Amount
		}
		totals[PaymentCash] -= sale.Change
	}
	for _, method := range paymentMethodOrder {
		if totals[method] != 0 {
			summary.ByMethod = append(summary.ByMethod, MethodTotal{Method: method, Amount: totals[method]})
		}
	}

	return summary
}
