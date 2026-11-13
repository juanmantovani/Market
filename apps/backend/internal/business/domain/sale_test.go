package domain

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

var (
	saleTime  = time.Date(2026, 11, 12, 15, 30, 0, 0, time.UTC)
	saleActor = Actor{ID: "u-cashier", Role: RoleCashier}
)

func saleItems() map[string]Item {
	return map[string]Item{
		"coca": {ID: "coca", ItemFields: ItemFields{Code: "BEB-1", Name: "Gaseosa cola", Unit: UnitPiece, SalePrice: 320_000, TaxRate: 2100}},
		"pan":  {ID: "pan", ItemFields: ItemFields{Code: "PAN-1", Name: "Pan francés", Unit: UnitKilogram, SalePrice: 280_000, TaxRate: 1050}},
	}
}

func saleCustomer() Customer {
	return Customer{ID: "c1", CustomerFields: CustomerFields{
		Name: "Ana Pérez", DocumentType: DocumentDNI, DocumentNumber: "28456789", TaxCondition: TaxFinalConsumer,
	}}
}

func saleSeller() *Seller {
	return &Seller{ID: "s1", SellerFields: SellerFields{Name: "Juan Pérez", Code: "JP"}}
}

// validSaleRequest sells 2 colas and 0,750 kg of bread with 10 % off:
// 640.000 + 210.000 − 21.000 = 829.000 cents, paid with 10.000 pesos cash.
func validSaleRequest() SaleRequest {
	return SaleRequest{
		IdempotencyKey: "key-1",
		CustomerID:     textPtr("c1"),
		SellerID:       textPtr("s1"),
		Lines: []SaleRequestLine{
			{ItemID: "coca", Quantity: 2_000},
			{ItemID: "pan", Quantity: 750, Discount: percentOff(1_000)},
		},
		Payments:      []Payment{{Method: PaymentCash, Amount: 1_000_000}},
		ExpectedTotal: 829_000,
	}
}

func mustSaleRequest(t *testing.T, request SaleRequest) SaleRequest {
	t.Helper()
	valid, err := NewSaleRequest(request)
	if err != nil {
		t.Fatalf("NewSaleRequest() error = %v", err)
	}
	return valid
}

func TestPaymentMethodIsValid(t *testing.T) {
	for _, method := range []PaymentMethod{PaymentCash, PaymentDebitCard, PaymentCreditCard, PaymentTransfer, PaymentQR} {
		if !method.IsValid() {
			t.Errorf("%q.IsValid() = false", method)
		}
	}
	if PaymentMethod("cheque").IsValid() {
		t.Error(`"cheque".IsValid() = true`)
	}
}

func TestNewSaleRequestNormalizes(t *testing.T) {
	request := SaleRequest{
		IdempotencyKey: "  key-1 ",
		CustomerID:     textPtr("  "),
		SellerID:       textPtr(" s1 "),
		Lines:          []SaleRequestLine{{ItemID: " coca ", Quantity: 1_000, Discount: percentOff(10_000)}},
		CartDiscount:   fixedOff(500),
		Payments:       []Payment{{Method: PaymentDebitCard, Amount: 320_000, Reference: textPtr(" Cupón 123 ")}, {Method: PaymentCash, Amount: 1, Reference: textPtr(" ")}},
	}

	valid := mustSaleRequest(t, request)

	if valid.IdempotencyKey != "key-1" || valid.CustomerID != nil || *valid.SellerID != "s1" || valid.Lines[0].ItemID != "coca" {
		t.Fatalf("request = %+v", valid)
	}
	if *valid.Payments[0].Reference != "Cupón 123" || valid.Payments[1].Reference != nil {
		t.Fatalf("payments = %+v", valid.Payments)
	}
	request.Lines[0].Quantity = 9
	*request.SellerID = "changed"
	request.Lines[0].Discount.Value, request.CartDiscount.Value = 1, 1
	if valid.Lines[0].Quantity != 1_000 || *valid.SellerID != "s1" || valid.Lines[0].Discount.Value != 10_000 || valid.CartDiscount.Value != 500 {
		t.Fatal("the validated request aliases the caller's data")
	}
}

func TestNewSaleRequestValidates(t *testing.T) {
	manyLines := make([]SaleRequestLine, MaxSaleLines+1)
	for i := range manyLines {
		manyLines[i] = SaleRequestLine{ItemID: string(rune('a'+i%26)) + strings.Repeat("x", i), Quantity: 1_000}
	}
	manyPayments := make([]Payment, MaxSalePayments+1)
	for i := range manyPayments {
		manyPayments[i] = Payment{Method: PaymentCash, Amount: 1}
	}

	tests := []struct {
		name   string
		change func(*SaleRequest)
		want   *Error
	}{
		{"no idempotency key", func(r *SaleRequest) { r.IdempotencyKey = "  " }, ErrIdempotencyKeyRequired},
		{"idempotency key too long", func(r *SaleRequest) { r.IdempotencyKey = strings.Repeat("k", MaxIdempotencyKeyLength+1) }, ErrIdempotencyKeyInvalid},
		{"no lines", func(r *SaleRequest) { r.Lines = nil }, ErrSaleLinesRequired},
		{"too many lines", func(r *SaleRequest) { r.Lines = manyLines }, ErrSaleTooManyLines},
		{"line without item", func(r *SaleRequest) { r.Lines[0].ItemID = " " }, ErrSaleItemRequired},
		{"repeated item", func(r *SaleRequest) { r.Lines[1].ItemID = " coca" }, ErrSaleItemRepeated},
		{"zero quantity", func(r *SaleRequest) { r.Lines[0].Quantity = 0 }, ErrSaleQuantityInvalid},
		{"negative quantity", func(r *SaleRequest) { r.Lines[0].Quantity = -1_000 }, ErrSaleQuantityInvalid},
		{"quantity over the maximum", func(r *SaleRequest) { r.Lines[0].Quantity = MaxStockQuantity + 1 }, ErrSaleQuantityInvalid},
		{"negative percent discount", func(r *SaleRequest) { r.Lines[0].Discount = percentOff(-1) }, ErrSaleDiscountInvalid},
		{"negative amount discount", func(r *SaleRequest) { r.Lines[0].Discount = fixedOff(-1) }, ErrSaleDiscountInvalid},
		{"amount discount over the maximum", func(r *SaleRequest) { r.Lines[0].Discount = fixedOff(int64(MaxSaleAmount) + 1) }, ErrSaleDiscountInvalid},
		{"unknown discount type", func(r *SaleRequest) { r.Lines[0].Discount = &Discount{Type: "gift", Value: 1} }, ErrSaleDiscountInvalid},
		{"invalid cart discount", func(r *SaleRequest) { r.CartDiscount = percentOff(PercentScale + 1) }, ErrSaleDiscountInvalid},
		{"percent discount over 100 %", func(r *SaleRequest) { r.Lines[0].Discount = percentOff(PercentScale + 1) }, ErrSaleDiscountInvalid},
		{"too many payments", func(r *SaleRequest) { r.Payments = manyPayments }, ErrSaleTooManyPayments},
		{"unknown payment method", func(r *SaleRequest) { r.Payments[0].Method = "cheque" }, ErrPaymentMethodInvalid},
		{"zero payment", func(r *SaleRequest) { r.Payments[0].Amount = 0 }, ErrPaymentAmountInvalid},
		{"payment over the maximum", func(r *SaleRequest) { r.Payments[0].Amount = MaxSaleAmount + 1 }, ErrPaymentAmountInvalid},
		{"payment reference too long", func(r *SaleRequest) {
			r.Payments[0].Reference = textPtr(strings.Repeat("r", MaxPaymentReferenceLength+1))
		}, ErrPaymentReferenceTooLong},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := validSaleRequest()
			test.change(&request)

			_, err := NewSaleRequest(request)

			assertDomainError(t, err, test.want)
		})
	}
}

func TestSaleRequestFingerprint(t *testing.T) {
	base := mustSaleRequest(t, validSaleRequest()).Fingerprint()
	if len(base) != 64 {
		t.Fatalf("Fingerprint() = %q, want a SHA-256 in hex", base)
	}

	sameContent := validSaleRequest()
	sameContent.IdempotencyKey = "another-key"
	if got := mustSaleRequest(t, sameContent).Fingerprint(); got != base {
		t.Fatal("the idempotency key changed the fingerprint")
	}

	changes := map[string]func(*SaleRequest){
		"customer":           func(r *SaleRequest) { r.CustomerID = nil },
		"seller":             func(r *SaleRequest) { r.SellerID = textPtr("s2") },
		"expected total":     func(r *SaleRequest) { r.ExpectedTotal++ },
		"line item":          func(r *SaleRequest) { r.Lines[0].ItemID = "agua" },
		"line quantity":      func(r *SaleRequest) { r.Lines[0].Quantity = 3_000 },
		"line discount":      func(r *SaleRequest) { r.Lines[1].Discount = nil },
		"line discount type": func(r *SaleRequest) { r.Lines[1].Discount = fixedOff(1_000) },
		"cart discount":      func(r *SaleRequest) { r.CartDiscount = percentOff(500) },
		"line order":         func(r *SaleRequest) { r.Lines[0], r.Lines[1] = r.Lines[1], r.Lines[0] },
		"payment method":     func(r *SaleRequest) { r.Payments[0].Method = PaymentQR },
		"payment amount":     func(r *SaleRequest) { r.Payments[0].Amount = 900_000 },
		"payment reference":  func(r *SaleRequest) { r.Payments[0].Reference = textPtr("Caja 2") },
		"another payment":    func(r *SaleRequest) { r.Payments = append(r.Payments, Payment{Method: PaymentCash, Amount: 1}) },
	}
	for name, change := range changes {
		request := validSaleRequest()
		change(&request)
		if mustSaleRequest(t, request).Fingerprint() == base {
			t.Errorf("changing the %s kept the fingerprint", name)
		}
	}
}

func TestNewSaleComputesLinesAndTotals(t *testing.T) {
	request := mustSaleRequest(t, validSaleRequest())

	sale, err := NewSale(request, saleCustomer(), saleSeller(), saleItems(), saleActor, saleTime)
	if err != nil {
		t.Fatalf("NewSale() error = %v", err)
	}

	coca, pan := sale.Lines[0], sale.Lines[1]
	if coca != (SaleLine{ItemID: "coca", Code: "BEB-1", Name: "Gaseosa cola", Unit: UnitPiece, TaxRate: 2100, Quantity: 2_000, UnitPrice: 320_000, LineTotal: 640_000}) {
		t.Errorf("coca line = %+v", coca)
	}
	if pan.Discount == nil || *pan.Discount != (Discount{Type: DiscountPercent, Value: 1_000}) {
		t.Errorf("bread discount = %+v", pan.Discount)
	}
	pan.Discount = nil
	if pan != (SaleLine{ItemID: "pan", Code: "PAN-1", Name: "Pan francés", Unit: UnitKilogram, TaxRate: 1050, Quantity: 750, UnitPrice: 280_000, DiscountAmount: 21_000, LineTotal: 189_000}) {
		t.Errorf("bread line = %+v", pan)
	}
	if sale.Subtotal != 850_000 || sale.DiscountTotal != 21_000 || sale.Total != 829_000 || sale.Change != 171_000 {
		t.Errorf("totals = subtotal %d, discount %d, total %d, change %d", sale.Subtotal, sale.DiscountTotal, sale.Total, sale.Change)
	}
	if sale.Customer != (SaleCustomer{ID: "c1", Name: "Ana Pérez", DocumentType: DocumentDNI, DocumentNumber: "28456789", TaxCondition: TaxFinalConsumer}) {
		t.Errorf("customer = %+v", sale.Customer)
	}
	if sale.Seller == nil || *sale.Seller != (SaleSeller{ID: "s1", Name: "Juan Pérez", Code: "JP"}) {
		t.Errorf("seller = %+v", sale.Seller)
	}
	if sale.ID != "" || sale.Number != 0 || sale.Status != SaleCompleted || sale.ActorID != "u-cashier" || !sale.CreatedAt.Equal(saleTime) ||
		sale.IdempotencyKey != "key-1" || sale.Fingerprint != request.Fingerprint() || len(sale.Payments) != 1 {
		t.Errorf("sale = %+v", sale)
	}
}

func TestNewSaleWithoutSeller(t *testing.T) {
	request := validSaleRequest()
	request.SellerID = nil

	sale, err := NewSale(mustSaleRequest(t, request), saleCustomer(), nil, saleItems(), saleActor, saleTime)
	if err != nil {
		t.Fatalf("NewSale() error = %v", err)
	}
	if sale.Seller != nil {
		t.Fatalf("seller = %+v, want none", sale.Seller)
	}
}

func TestNewSaleRoundsEachLineToTheCent(t *testing.T) {
	tests := []struct {
		name                        string
		price                       Money
		quantity                    Quantity
		discount                    int64
		wantDiscount, wantLineTotal Money
	}{
		{name: "whole units", price: 120_000, quantity: 6_000, wantLineTotal: 720_000},
		{name: "fraction rounds half up", price: 99_999, quantity: 333, wantLineTotal: 33_300},
		{name: "fraction rounds down", price: 1_001, quantity: 1, wantLineTotal: 1},
		{name: "discount rounds half up", price: 33_300, quantity: 1_000, discount: 1_250, wantDiscount: 4_163, wantLineTotal: 29_137},
		{name: "full discount", price: 50_000, quantity: 3_000, discount: PercentScale, wantDiscount: 150_000, wantLineTotal: 0},
		{name: "a cent at half a gram", price: 1, quantity: 500, wantLineTotal: 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			items := map[string]Item{"x": {ID: "x", ItemFields: ItemFields{Unit: UnitKilogram, SalePrice: test.price}}}
			request := SaleRequest{
				IdempotencyKey: "k",
				Lines:          []SaleRequestLine{{ItemID: "x", Quantity: test.quantity, Discount: percentOff(test.discount)}},
				ExpectedTotal:  test.wantLineTotal,
			}
			if test.wantLineTotal > 0 {
				request.Payments = []Payment{{Method: PaymentCash, Amount: test.wantLineTotal}}
			}

			sale, err := NewSale(mustSaleRequest(t, request), saleCustomer(), nil, items, saleActor, saleTime)
			if err != nil {
				t.Fatalf("NewSale() error = %v", err)
			}
			line := sale.Lines[0]
			if line.DiscountAmount != test.wantDiscount || line.LineTotal != test.wantLineTotal {
				t.Fatalf("line = discount %d, total %d; want %d, %d", line.DiscountAmount, line.LineTotal, test.wantDiscount, test.wantLineTotal)
			}
			if sale.Subtotal-sale.DiscountTotal != sale.Total || sale.Total != line.LineTotal {
				t.Fatalf("totals = %d − %d = %d", sale.Subtotal, sale.DiscountTotal, sale.Total)
			}
		})
	}
}

func TestNewSalePayments(t *testing.T) {
	cash := func(amount Money) Payment { return Payment{Method: PaymentCash, Amount: amount} }
	debit := func(amount Money) Payment { return Payment{Method: PaymentDebitCard, Amount: amount} }
	transfer := func(amount Money) Payment { return Payment{Method: PaymentTransfer, Amount: amount} }

	tests := []struct {
		name       string
		payments   []Payment
		wantChange Money
		wantErr    *Error
	}{
		{name: "exact cash", payments: []Payment{cash(829_000)}},
		{name: "cash with change", payments: []Payment{cash(1_000_000)}, wantChange: 171_000},
		{name: "exact debit", payments: []Payment{debit(829_000)}},
		{name: "debit plus cash with change", payments: []Payment{debit(500_000), cash(500_000)}, wantChange: 171_000},
		{name: "two non-cash payments", payments: []Payment{debit(429_000), transfer(400_000)}},
		{name: "non-cash over the total", payments: []Payment{debit(900_000)}, wantErr: ErrSaleNonCashPaymentExceedsTotal},
		{name: "non-cash over the total with cash", payments: []Payment{debit(500_000), transfer(400_000), cash(100)}, wantErr: ErrSaleNonCashPaymentExceedsTotal},
		{name: "not enough", payments: []Payment{cash(800_000), debit(29_000 - 1)}, wantErr: ErrSalePaymentInsufficient},
		{name: "no payments", payments: nil, wantErr: ErrSalePaymentInsufficient},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := validSaleRequest()
			request.Payments = test.payments

			sale, err := NewSale(mustSaleRequest(t, request), saleCustomer(), saleSeller(), saleItems(), saleActor, saleTime)

			if test.wantErr != nil {
				assertDomainError(t, err, test.wantErr)
				return
			}
			if err != nil {
				t.Fatalf("NewSale() error = %v", err)
			}
			if sale.Change != test.wantChange || len(sale.Payments) != len(test.payments) {
				t.Fatalf("change = %d, payments = %+v; want change %d", sale.Change, sale.Payments, test.wantChange)
			}
		})
	}
}

func TestNewSaleWithAZeroTotalNeedsNoPayment(t *testing.T) {
	request := validSaleRequest()
	request.Lines = []SaleRequestLine{{ItemID: "coca", Quantity: 1_000, Discount: percentOff(PercentScale)}}
	request.Payments = nil
	request.ExpectedTotal = 0

	sale, err := NewSale(mustSaleRequest(t, request), saleCustomer(), nil, saleItems(), saleActor, saleTime)
	if err != nil {
		t.Fatalf("NewSale() error = %v", err)
	}
	if sale.Total != 0 || sale.DiscountTotal != 320_000 || sale.Change != 0 {
		t.Fatalf("sale = %+v", sale)
	}
}

func TestNewSaleErrors(t *testing.T) {
	tests := []struct {
		name    string
		change  func(*SaleRequest, map[string]Item)
		wantErr *Error
	}{
		{
			name:    "the total the cashier saw differs",
			change:  func(r *SaleRequest, _ map[string]Item) { r.ExpectedTotal = 828_999 },
			wantErr: ErrSaleTotalMismatch,
		},
		{
			name: "a price changed while building the cart",
			change: func(_ *SaleRequest, items map[string]Item) {
				coca := items["coca"]
				coca.SalePrice = 330_000
				items["coca"] = coca
			},
			wantErr: ErrSaleTotalMismatch,
		},
		{
			name: "an item is discontinued",
			change: func(_ *SaleRequest, items map[string]Item) {
				pan := items["pan"]
				pan.Discontinued = true
				items["pan"] = pan
			},
			wantErr: ErrSaleItemDiscontinued,
		},
		{
			name:    "an item is missing",
			change:  func(_ *SaleRequest, items map[string]Item) { delete(items, "pan") },
			wantErr: ErrSaleItemNotFound,
		},
		{
			name:    "a piece sold in fractions",
			change:  func(r *SaleRequest, _ map[string]Item) { r.Lines[0].Quantity = 1_500 },
			wantErr: ErrSaleQuantityNotWhole,
		},
		{
			name: "a line too expensive to price",
			change: func(_ *SaleRequest, items map[string]Item) {
				coca := items["coca"]
				coca.SalePrice = Money(1 << 62)
				items["coca"] = coca
			},
			wantErr: ErrAmountOutOfRange,
		},
		{
			name: "a total over the maximum",
			change: func(r *SaleRequest, items map[string]Item) {
				coca := items["coca"]
				coca.SalePrice = MaxItemAmount
				items["coca"] = coca
				r.Lines[0].Quantity = MaxStockQuantity
			},
			wantErr: ErrAmountOutOfRange,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request, items := validSaleRequest(), saleItems()
			test.change(&request, items)

			_, err := NewSale(mustSaleRequest(t, request), saleCustomer(), saleSeller(), items, saleActor, saleTime)

			assertDomainError(t, err, test.wantErr)
		})
	}
}

func TestNewSaleRejectsPaymentsOverTheMaximum(t *testing.T) {
	request := validSaleRequest()
	request.Payments = []Payment{{Method: PaymentCash, Amount: MaxSaleAmount}, {Method: PaymentCash, Amount: MaxSaleAmount}}

	_, err := NewSale(mustSaleRequest(t, request), saleCustomer(), nil, saleItems(), saleActor, saleTime)

	assertDomainError(t, err, ErrAmountOutOfRange)
}

func TestSaleReplays(t *testing.T) {
	request := mustSaleRequest(t, validSaleRequest())
	sale, err := NewSale(request, saleCustomer(), saleSeller(), saleItems(), saleActor, saleTime)
	if err != nil {
		t.Fatalf("NewSale() error = %v", err)
	}

	if err := sale.Replays(request); err != nil {
		t.Fatalf("Replays(same request) error = %v", err)
	}
	other := validSaleRequest()
	other.Lines[0].Quantity = 3_000
	assertDomainError(t, sale.Replays(mustSaleRequest(t, other)), ErrIdempotencyKeyConflict)
}

func TestSaleReference(t *testing.T) {
	if got := (Sale{Number: 12}).Reference(); got != "Venta 12" {
		t.Fatalf("Reference() = %q", got)
	}
}

func TestNewSaleLineGuardsTheDiscountAgainstOverflow(t *testing.T) {
	item := Item{ID: "x", ItemFields: ItemFields{Unit: UnitPiece, SalePrice: MaxItemAmount}}

	_, err := newSaleLine(item, SaleRequestLine{ItemID: "x", Quantity: 1_000, Discount: percentOff(1 << 62)})

	assertDomainError(t, err, ErrAmountOutOfRange)
}

func completedSale(t *testing.T) Sale {
	t.Helper()
	sale, err := NewSale(mustSaleRequest(t, validSaleRequest()), saleCustomer(), saleSeller(), saleItems(), saleActor, saleTime)
	if err != nil {
		t.Fatalf("NewSale() error = %v", err)
	}
	sale.ID, sale.Number = "sale-1", 12
	return sale
}

func TestSaleStatusIsValid(t *testing.T) {
	if !SaleCompleted.IsValid() || !SaleVoided.IsValid() || SaleStatus("pending").IsValid() {
		t.Fatal("IsValid() accepts the wrong statuses")
	}
}

func TestSaleVoid(t *testing.T) {
	sale := completedSale(t)
	manager := Actor{ID: "u-manager", Role: RoleManager}
	later := saleTime.Add(time.Hour)

	voided, err := sale.Void("  El cliente se arrepintió ", manager, later)
	if err != nil {
		t.Fatalf("Void() error = %v", err)
	}

	if voided.Status != SaleVoided || voided.VoidedAt == nil || !voided.VoidedAt.Equal(later) ||
		*voided.VoidReason != "El cliente se arrepintió" || *voided.VoidedBy != "u-manager" {
		t.Fatalf("voided = %+v", voided)
	}
	if voided.ID != sale.ID || voided.Number != 12 || voided.Total != sale.Total || voided.ActorID != "u-cashier" {
		t.Fatalf("voiding changed the sale itself: %+v", voided)
	}
	if sale.Status != SaleCompleted || sale.VoidedAt != nil {
		t.Fatal("Void() changed the receiver")
	}

	_, err = voided.Void("Otra vez", manager, later)
	assertDomainError(t, err, ErrSaleAlreadyVoided)
	_, err = sale.Void("  ", manager, later)
	assertDomainError(t, err, ErrSaleVoidReasonRequired)
	_, err = sale.Void(strings.Repeat("m", MaxSaleVoidReasonLength+1), manager, later)
	assertDomainError(t, err, ErrSaleVoidReasonTooLong)
}

func TestNewSaleFilter(t *testing.T) {
	valid, err := NewSaleFilter(SaleFilter{From: saleTime, To: saleTime.Add(time.Hour), CustomerID: " c1 ", SellerID: " s1 ", Status: SaleVoided})
	if err != nil {
		t.Fatalf("NewSaleFilter() error = %v", err)
	}
	if valid.CustomerID != "c1" || valid.SellerID != "s1" || valid.Status != SaleVoided {
		t.Fatalf("filter = %+v", valid)
	}
	if _, err := NewSaleFilter(SaleFilter{}); err != nil {
		t.Fatalf("NewSaleFilter(empty) error = %v", err)
	}
	if _, err := NewSaleFilter(SaleFilter{From: saleTime}); err != nil {
		t.Fatalf("NewSaleFilter(only from) error = %v", err)
	}

	_, err = NewSaleFilter(SaleFilter{Status: "pending"})
	assertDomainError(t, err, ErrSaleStatusInvalid)
	_, err = NewSaleFilter(SaleFilter{From: saleTime, To: saleTime})
	assertDomainError(t, err, ErrSaleDateRangeInvalid)
	_, err = NewSaleFilter(SaleFilter{From: saleTime, To: saleTime.Add(-time.Second)})
	assertDomainError(t, err, ErrSaleDateRangeInvalid)
}

func TestSaleFilterMatches(t *testing.T) {
	sale := completedSale(t)
	withoutSeller := sale
	withoutSeller.Seller = nil

	tests := []struct {
		name   string
		filter SaleFilter
		sale   Sale
		want   bool
	}{
		{name: "no filter", filter: SaleFilter{}, sale: sale, want: true},
		{name: "from includes the instant", filter: SaleFilter{From: saleTime}, sale: sale, want: true},
		{name: "from after the sale", filter: SaleFilter{From: saleTime.Add(time.Nanosecond)}, sale: sale, want: false},
		{name: "to excludes the instant", filter: SaleFilter{To: saleTime}, sale: sale, want: false},
		{name: "to after the sale", filter: SaleFilter{To: saleTime.Add(time.Nanosecond)}, sale: sale, want: true},
		{name: "same customer", filter: SaleFilter{CustomerID: "c1"}, sale: sale, want: true},
		{name: "other customer", filter: SaleFilter{CustomerID: "c2"}, sale: sale, want: false},
		{name: "same seller", filter: SaleFilter{SellerID: "s1"}, sale: sale, want: true},
		{name: "other seller", filter: SaleFilter{SellerID: "s2"}, sale: sale, want: false},
		{name: "a seller, the sale has none", filter: SaleFilter{SellerID: "s1"}, sale: withoutSeller, want: false},
		{name: "same status", filter: SaleFilter{Status: SaleCompleted}, sale: sale, want: true},
		{name: "other status", filter: SaleFilter{Status: SaleVoided}, sale: sale, want: false},
	}
	for _, test := range tests {
		if got := test.filter.Matches(test.sale); got != test.want {
			t.Errorf("%s: Matches() = %v, want %v", test.name, got, test.want)
		}
	}
}

func percentOff(hundredths int64) *Discount {
	return &Discount{Type: DiscountPercent, Value: hundredths}
}

func fixedOff(cents int64) *Discount {
	return &Discount{Type: DiscountAmount, Value: cents}
}

func TestNewSaleWithAFixedLineDiscount(t *testing.T) {
	request := validSaleRequest()
	request.Lines[0].Discount = fixedOff(40_000)
	request.ExpectedTotal = 789_000

	sale, err := NewSale(mustSaleRequest(t, request), saleCustomer(), nil, saleItems(), saleActor, saleTime)
	if err != nil {
		t.Fatalf("NewSale() error = %v", err)
	}
	coca := sale.Lines[0]
	if coca.DiscountAmount != 40_000 || coca.LineTotal != 600_000 || coca.Amount() != 640_000 || *coca.Discount != *fixedOff(40_000) {
		t.Fatalf("coca line = %+v", coca)
	}
	if sale.Subtotal != 850_000 || sale.DiscountTotal != 61_000 || sale.Total != 789_000 {
		t.Fatalf("totals = %d − %d = %d", sale.Subtotal, sale.DiscountTotal, sale.Total)
	}

	request.Lines[0].Discount = fixedOff(640_000)
	request.ExpectedTotal = 189_000
	if whole, err := NewSale(mustSaleRequest(t, request), saleCustomer(), nil, saleItems(), saleActor, saleTime); err != nil || whole.Lines[0].LineTotal != 0 {
		t.Fatalf("a discount of the whole line = %+v, %v", whole.Lines, err)
	}
	request.Lines[0].Discount = fixedOff(640_001)
	_, err = NewSale(mustSaleRequest(t, request), saleCustomer(), nil, saleItems(), saleActor, saleTime)
	assertDomainError(t, err, ErrSaleDiscountExceedsLine)
}

func TestNewSaleWithACartDiscount(t *testing.T) {
	tests := []struct {
		name       string
		discount   *Discount
		wantShares []Money
		wantOff    Money
	}{
		// 640.000 + 189.000 after the bread's 10 %: 829.000.
		{name: "a percent, on what is left after the line discounts", discount: percentOff(1_000), wantShares: []Money{64_000, 18_900}, wantOff: 82_900},
		{name: "a fixed amount, spread by weight", discount: fixedOff(100_000), wantShares: []Money{77_201, 22_799}, wantOff: 100_000},
		{name: "everything", discount: fixedOff(829_000), wantShares: []Money{640_000, 189_000}, wantOff: 829_000},
		{name: "nothing", discount: fixedOff(0), wantShares: []Money{0, 0}, wantOff: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := validSaleRequest()
			request.CartDiscount = test.discount
			request.ExpectedTotal = 829_000 - test.wantOff
			request.Payments = nil
			if request.ExpectedTotal > 0 {
				request.Payments = []Payment{{Method: PaymentCash, Amount: request.ExpectedTotal}}
			}

			sale, err := NewSale(mustSaleRequest(t, request), saleCustomer(), nil, saleItems(), saleActor, saleTime)
			if err != nil {
				t.Fatalf("NewSale() error = %v", err)
			}

			var lineTotals Money
			for i, line := range sale.Lines {
				if line.CartDiscountAmount != test.wantShares[i] {
					t.Errorf("line %d share = %d, want %d", i, line.CartDiscountAmount, test.wantShares[i])
				}
				lineTotals += line.LineTotal
			}
			if sale.CartDiscountAmount != test.wantOff || *sale.CartDiscount != *test.discount ||
				sale.DiscountTotal != 21_000+test.wantOff || sale.Subtotal != 850_000 || sale.Total != lineTotals {
				t.Fatalf("sale = cart %d (%+v), discounts %d, subtotal %d, total %d, lines %d",
					sale.CartDiscountAmount, sale.CartDiscount, sale.DiscountTotal, sale.Subtotal, sale.Total, lineTotals)
			}
		})
	}
}

func TestNewSaleRefusesACartDiscountOverTheTotal(t *testing.T) {
	request := validSaleRequest()
	request.CartDiscount = fixedOff(829_001)

	_, err := NewSale(mustSaleRequest(t, request), saleCustomer(), nil, saleItems(), saleActor, saleTime)

	assertDomainError(t, err, ErrSaleDiscountExceedsTotal)
}

func TestProrate(t *testing.T) {
	tests := []struct {
		name    string
		total   Money
		weights []Money
		want    []Money
	}{
		{name: "nothing to spread", total: 0, weights: []Money{5, 7}, want: []Money{0, 0}},
		{name: "no weight", total: 0, weights: []Money{0, 0}, want: []Money{0, 0}},
		{name: "exact shares", total: 30, weights: []Money{200, 100}, want: []Money{20, 10}},
		{name: "cents go to the largest fractions", total: 6, weights: []Money{3, 3, 1}, want: []Money{3, 2, 1}},
		{name: "the first line wins a tie", total: 1, weights: []Money{1, 1}, want: []Money{1, 0}},
		{name: "a line without amount gets nothing", total: 5, weights: []Money{5, 0}, want: []Money{5, 0}},
		{name: "beyond int64 before dividing", total: MaxSaleAmount, weights: []Money{MaxSaleAmount / 2, MaxSaleAmount / 2}, want: []Money{MaxSaleAmount / 2, MaxSaleAmount / 2}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := prorate(test.total, test.weights)

			var sum Money
			for i, share := range got {
				sum += share
				if share > test.weights[i] {
					t.Errorf("share %d = %d, over its weight %d", i, share, test.weights[i])
				}
			}
			if fmt.Sprint(got) != fmt.Sprint(test.want) || sum != test.total {
				t.Fatalf("prorate(%d, %v) = %v, want %v", test.total, test.weights, got, test.want)
			}
		})
	}
}

func TestDiscountString(t *testing.T) {
	var none *Discount
	if none.String() != "none" || percentOff(1_000).String() != "percent:1000" || fixedOff(500).String() != "amount:500" {
		t.Fatal("String() doesn't tell discounts apart")
	}
}
