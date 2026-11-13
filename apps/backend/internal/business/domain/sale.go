package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaxSaleLines              = 100
	MaxSalePayments           = 10
	MaxPaymentReferenceLength = 60
	MaxIdempotencyKeyLength   = 100
	MaxSaleVoidReasonLength   = 200

	// MaxSaleAmount caps the total of a sale and the sum of its payments at
	// a hundred thousand million pesos, so adding lines never overflows.
	MaxSaleAmount = 100 * MaxItemAmount
)

// PaymentMethod is how part of a sale is paid. Only cash can exceed the
// total: the difference is the change.
type PaymentMethod string

const (
	PaymentCash       PaymentMethod = "cash"
	PaymentDebitCard  PaymentMethod = "debit_card"
	PaymentCreditCard PaymentMethod = "credit_card"
	PaymentTransfer   PaymentMethod = "transfer"
	PaymentQR         PaymentMethod = "qr"
)

func (method PaymentMethod) IsValid() bool {
	switch method {
	case PaymentCash, PaymentDebitCard, PaymentCreditCard, PaymentTransfer, PaymentQR:
		return true
	default:
		return false
	}
}

type SaleStatus string

const (
	SaleCompleted SaleStatus = "completed"
	SaleVoided    SaleStatus = "voided"
)

var (
	ErrSaleNotFound                   = &Error{Kind: KindNotFound, Code: "sale_not_found", Message: "La venta no existe"}
	ErrIdempotencyKeyRequired         = &Error{Kind: KindInvalid, Code: "idempotency_key_required", Message: "No se pudo identificar la operación: volver a intentar"}
	ErrIdempotencyKeyInvalid          = &Error{Kind: KindInvalid, Code: "idempotency_key_invalid", Message: "No se pudo identificar la operación: volver a intentar"}
	ErrIdempotencyKeyConflict         = &Error{Kind: KindConflict, Code: "idempotency_key_conflict", Message: "La operación ya se registró con otros datos: revisar la venta y volver a cobrar"}
	ErrSaleLinesRequired              = &Error{Kind: KindInvalid, Code: "sale_lines_required", Message: "Agregar al menos un artículo"}
	ErrSaleTooManyLines               = &Error{Kind: KindInvalid, Code: "sale_too_many_lines", Message: "La venta admite hasta 100 artículos"}
	ErrSaleItemRequired               = &Error{Kind: KindInvalid, Code: "sale_item_required", Message: "Elegir el artículo"}
	ErrSaleItemRepeated               = &Error{Kind: KindInvalid, Code: "sale_item_repeated", Message: "Un artículo aparece en más de una línea: sumar las cantidades en una sola"}
	ErrSaleItemNotFound               = &Error{Kind: KindInvalid, Code: "sale_item_not_found", Message: "Uno de los artículos de la venta ya no existe"}
	ErrSaleQuantityInvalid            = &Error{Kind: KindInvalid, Code: "sale_quantity_invalid", Message: "La cantidad tiene que ser mayor a cero y de hasta un millón de unidades"}
	ErrSaleQuantityNotWhole           = &Error{Kind: KindInvalid, Code: "sale_quantity_not_whole", Message: "El artículo se vende por unidad: la cantidad tiene que ser entera"}
	ErrSaleCustomerNotFound           = &Error{Kind: KindInvalid, Code: "sale_customer_not_found", Message: "El cliente elegido no existe"}
	ErrSaleSellerNotFound             = &Error{Kind: KindInvalid, Code: "sale_seller_not_found", Message: "El vendedor elegido no existe"}
	ErrSaleTotalMismatch              = &Error{Kind: KindConflict, Code: "sale_total_mismatch", Message: "El total cambió mientras se armaba la venta: revisar el carrito antes de cobrar"}
	ErrSaleTooManyPayments            = &Error{Kind: KindInvalid, Code: "sale_too_many_payments", Message: "La venta admite hasta 10 pagos"}
	ErrPaymentMethodInvalid           = &Error{Kind: KindInvalid, Code: "payment_method_invalid", Message: "El medio de pago no es válido"}
	ErrPaymentAmountInvalid           = &Error{Kind: KindInvalid, Code: "payment_amount_invalid", Message: "El importe de cada pago tiene que ser mayor a cero"}
	ErrPaymentReferenceTooLong        = &Error{Kind: KindInvalid, Code: "payment_reference_too_long", Message: "La referencia del pago admite hasta 60 caracteres"}
	ErrSalePaymentInsufficient        = &Error{Kind: KindInvalid, Code: "sale_payment_insufficient", Message: "Los pagos no cubren el total de la venta"}
	ErrSaleNonCashPaymentExceedsTotal = &Error{Kind: KindInvalid, Code: "sale_non_cash_payment_exceeds_total", Message: "Solo el efectivo puede superar el total: ajustar los otros medios de pago"}
	ErrSaleAlreadyVoided              = &Error{Kind: KindConflict, Code: "sale_already_voided", Message: "La venta ya está anulada"}
	ErrSaleNotVoided                  = &Error{Kind: KindConflict, Code: "sale_not_voided", Message: "La venta no está anulada"}
	ErrSaleVoidReasonRequired         = &Error{Kind: KindInvalid, Code: "sale_void_reason_required", Message: "Ingresar el motivo de la anulación"}
	ErrSaleVoidReasonTooLong          = &Error{Kind: KindInvalid, Code: "sale_void_reason_too_long", Message: "El motivo admite hasta 200 caracteres"}
	ErrSaleStatusInvalid              = &Error{Kind: KindInvalid, Code: "sale_status_invalid", Message: "El estado de la venta no es válido"}
	ErrSaleDateInvalid                = &Error{Kind: KindInvalid, Code: "sale_date_invalid", Message: "La fecha del filtro no es válida"}
	ErrSaleDateRangeInvalid           = &Error{Kind: KindInvalid, Code: "sale_date_range_invalid", Message: "La fecha de inicio tiene que ser anterior a la de fin"}
)

func (status SaleStatus) IsValid() bool {
	return status == SaleCompleted || status == SaleVoided
}

// SaleRequestLine is one item the cashier wants to sell, with an optional
// discount for this line only.
type SaleRequestLine struct {
	ItemID   string
	Quantity Quantity
	Discount *Discount
}

// SaleRequest is what the counter asks to sell, before reading any stored
// data. ExpectedTotal is the total the cashier saw: the sale is refused if
// the server computes another one.
type SaleRequest struct {
	IdempotencyKey string
	// CustomerID is optional: a sale without customer goes to Consumidor
	// Final.
	CustomerID *string
	SellerID   *string
	Lines      []SaleRequestLine
	// CartDiscount, optional, is the general discount of the whole sale. It
	// applies after the line discounts.
	CartDiscount  *Discount
	Payments      []Payment
	ExpectedTotal Money
}

// NewSaleRequest validates what a sale request must satisfy before looking
// at any stored data. The result never aliases the caller's pointers or
// slices.
func NewSaleRequest(request SaleRequest) (SaleRequest, error) {
	key := strings.TrimSpace(request.IdempotencyKey)
	if key == "" {
		return SaleRequest{}, ErrIdempotencyKeyRequired
	}
	if utf8.RuneCountInString(key) > MaxIdempotencyKeyLength {
		return SaleRequest{}, ErrIdempotencyKeyInvalid
	}

	lines, err := validateSaleRequestLines(request.Lines)
	if err != nil {
		return SaleRequest{}, err
	}
	cartDiscount, err := validateDiscount(request.CartDiscount)
	if err != nil {
		return SaleRequest{}, err
	}
	payments, err := validatePayments(request.Payments)
	if err != nil {
		return SaleRequest{}, err
	}

	return SaleRequest{
		IdempotencyKey: key,
		CustomerID:     optionalID(request.CustomerID),
		SellerID:       optionalID(request.SellerID),
		Lines:          lines,
		CartDiscount:   cartDiscount,
		Payments:       payments,
		ExpectedTotal:  request.ExpectedTotal,
	}, nil
}

func validateSaleRequestLines(lines []SaleRequestLine) ([]SaleRequestLine, error) {
	if len(lines) == 0 {
		return nil, ErrSaleLinesRequired
	}
	if len(lines) > MaxSaleLines {
		return nil, ErrSaleTooManyLines
	}

	valid := make([]SaleRequestLine, 0, len(lines))
	seen := make(map[string]bool, len(lines))
	for _, line := range lines {
		itemID := strings.TrimSpace(line.ItemID)
		if itemID == "" {
			return nil, ErrSaleItemRequired
		}
		if seen[itemID] {
			return nil, ErrSaleItemRepeated
		}
		seen[itemID] = true
		if line.Quantity <= 0 || line.Quantity > MaxStockQuantity {
			return nil, ErrSaleQuantityInvalid
		}
		discount, err := validateDiscount(line.Discount)
		if err != nil {
			return nil, err
		}
		valid = append(valid, SaleRequestLine{ItemID: itemID, Quantity: line.Quantity, Discount: discount})
	}

	return valid, nil
}

func validatePayments(payments []Payment) ([]Payment, error) {
	if len(payments) > MaxSalePayments {
		return nil, ErrSaleTooManyPayments
	}

	valid := make([]Payment, 0, len(payments))
	for _, payment := range payments {
		if !payment.Method.IsValid() {
			return nil, ErrPaymentMethodInvalid
		}
		if payment.Amount <= 0 || payment.Amount > MaxSaleAmount {
			return nil, ErrPaymentAmountInvalid
		}
		reference, err := optionalText(payment.Reference, MaxPaymentReferenceLength, ErrPaymentReferenceTooLong)
		if err != nil {
			return nil, err
		}
		valid = append(valid, Payment{Method: payment.Method, Amount: payment.Amount, Reference: reference})
	}

	return valid, nil
}

// Fingerprint identifies the content of a validated request, so a retry
// with the same idempotency key is told apart from a different sale that
// reused it.
func (request SaleRequest) Fingerprint() string {
	hash := sha256.New()
	fmt.Fprintf(hash, "customer=%q seller=%q expected=%d cart-discount=%s\n",
		valueOr(request.CustomerID), valueOr(request.SellerID), request.ExpectedTotal, request.CartDiscount)
	for _, line := range request.Lines {
		fmt.Fprintf(hash, "line=%q %d %s\n", line.ItemID, line.Quantity, line.Discount)
	}
	for _, payment := range request.Payments {
		fmt.Fprintf(hash, "payment=%q %d %q\n", payment.Method, payment.Amount, valueOr(payment.Reference))
	}

	return hex.EncodeToString(hash.Sum(nil))
}

func valueOr(text *string) string {
	if text == nil {
		return ""
	}

	return *text
}

// SaleCustomer is who the sale was made to, copied when it was made so
// editing the customer later doesn't change past sales.
type SaleCustomer struct {
	ID             string
	Name           string
	DocumentType   DocumentType
	DocumentNumber string
	TaxCondition   TaxCondition
}

// SaleSeller is who attended the sale, copied like the customer.
type SaleSeller struct {
	ID   string
	Name string
	Code string
}

// SaleLine is one item sold. Code, name, unit, tax rate and price are
// copied from the item, so editing it later doesn't change the sale.
type SaleLine struct {
	ItemID   string
	Code     string
	Name     string
	Unit     Unit
	TaxRate  Percent
	Quantity Quantity
	// UnitPrice includes IVA, like the item's sale price.
	UnitPrice Money
	// Discount is the line's own discount as it was asked for, and
	// DiscountAmount what it took off.
	Discount       *Discount
	DiscountAmount Money
	// CartDiscountAmount is this line's share of the sale's general
	// discount.
	CartDiscountAmount Money
	// LineTotal is UnitPrice × Quantity minus both discounts.
	LineTotal Money
}

// Amount is what the line costs before any discount.
func (line SaleLine) Amount() Money {
	return line.LineTotal + line.DiscountAmount + line.CartDiscountAmount
}

type Payment struct {
	Method PaymentMethod
	Amount Money
	// Reference is optional: a card voucher, a transfer number.
	Reference *string
}

// Sale is a confirmed sale. It is never edited or deleted: voiding it
// compensates its stock movements.
type Sale struct {
	ID string
	// Number is correlative and assigned by the repository.
	Number   int64
	Customer SaleCustomer
	// Seller is nil for a sale nobody in particular attended.
	Seller *SaleSeller
	Lines  []SaleLine
	// Subtotal adds the lines before any discount.
	Subtotal Money
	// CartDiscount is the general discount as it was asked for, and
	// CartDiscountAmount what it took off, spread over the lines.
	CartDiscount       *Discount
	CartDiscountAmount Money
	// DiscountTotal adds the line discounts and the general one.
	DiscountTotal Money
	Total         Money
	Payments      []Payment
	// Change is what cash paid beyond the total.
	Change         Money
	Status         SaleStatus
	ActorID        string
	IdempotencyKey string
	// Fingerprint is the request's, to answer a retry with this same sale.
	Fingerprint string
	CreatedAt   time.Time
	// VoidedAt, VoidReason and VoidedBy are set once the sale is voided.
	VoidedAt   *time.Time
	VoidReason *string
	VoidedBy   *string
}

// NewSale computes a sale from a validated request and the stored records
// it names: items holds every item of the request by ID and seller is nil
// when the request names none. Prices and totals come from the items, never
// from the counter, and the result must match ExpectedTotal.
func NewSale(request SaleRequest, customer Customer, seller *Seller, items map[string]Item, actor Actor, now time.Time) (Sale, error) {
	sale := Sale{
		Customer: SaleCustomer{
			ID:             customer.ID,
			Name:           customer.Name,
			DocumentType:   customer.DocumentType,
			DocumentNumber: customer.DocumentNumber,
			TaxCondition:   customer.TaxCondition,
		},
		Lines:          make([]SaleLine, 0, len(request.Lines)),
		Status:         SaleCompleted,
		ActorID:        actor.ID,
		IdempotencyKey: request.IdempotencyKey,
		Fingerprint:    request.Fingerprint(),
		CreatedAt:      now,
	}
	if seller != nil {
		sale.Seller = &SaleSeller{ID: seller.ID, Name: seller.Name, Code: seller.Code}
	}

	var afterLineDiscounts Money
	for _, requested := range request.Lines {
		item, ok := items[requested.ItemID]
		if !ok {
			return Sale{}, ErrSaleItemNotFound
		}
		line, err := newSaleLine(item, requested)
		if err != nil {
			return Sale{}, err
		}
		sale.Lines = append(sale.Lines, line)
		sale.Subtotal += line.Amount()
		sale.DiscountTotal += line.DiscountAmount
		afterLineDiscounts += line.LineTotal
		if sale.Subtotal > MaxSaleAmount {
			return Sale{}, ErrAmountOutOfRange
		}
	}
	if err := sale.applyCartDiscount(request.CartDiscount, afterLineDiscounts); err != nil {
		return Sale{}, err
	}
	sale.Total = sale.Subtotal - sale.DiscountTotal
	if sale.Total != request.ExpectedTotal {
		return Sale{}, ErrSaleTotalMismatch
	}

	change, err := changeFor(sale.Total, request.Payments)
	if err != nil {
		return Sale{}, err
	}
	sale.Payments = request.Payments
	sale.Change = change

	return sale, nil
}

// newSaleLine prices one line: the amount is rounded to the cent, then its
// discount (a percent rounded to the cent, or a fixed amount up to the line's
// amount) is taken from it, so the lines of a ticket always add up to its
// total.
func newSaleLine(item Item, requested SaleRequestLine) (SaleLine, error) {
	if !item.Unit.Accepts(requested.Quantity) {
		return SaleLine{}, ErrSaleQuantityNotWhole
	}
	amount, err := item.SalePrice.Times(requested.Quantity)
	if err != nil {
		return SaleLine{}, err
	}
	discount, err := requested.Discount.amountOff(amount, ErrSaleDiscountExceedsLine)
	if err != nil {
		return SaleLine{}, err
	}

	return SaleLine{
		ItemID:         item.ID,
		Code:           item.Code,
		Name:           item.Name,
		Unit:           item.Unit,
		TaxRate:        item.TaxRate,
		Quantity:       requested.Quantity,
		UnitPrice:      item.SalePrice,
		Discount:       requested.Discount,
		DiscountAmount: discount,
		LineTotal:      amount - discount,
	}, nil
}

// applyCartDiscount takes the general discount off what the lines add up to
// after their own discounts, and spreads it over the lines in proportion to
// that (prorate), so each line knows its share and the total stays the sum
// of the lines.
func (sale *Sale) applyCartDiscount(discount *Discount, afterLineDiscounts Money) error {
	off, err := discount.amountOff(afterLineDiscounts, ErrSaleDiscountExceedsTotal)
	if err != nil {
		return err
	}

	weights := make([]Money, len(sale.Lines))
	for i, line := range sale.Lines {
		weights[i] = line.LineTotal
	}
	for i, share := range prorate(off, weights) {
		sale.Lines[i].CartDiscountAmount = share
		sale.Lines[i].LineTotal -= share
	}
	sale.CartDiscount = discount
	sale.CartDiscountAmount = off
	sale.DiscountTotal += off

	return nil
}

// changeFor checks that payments cover total and returns the change. Card,
// transfer and QR payments can't add up to more than the total: only cash
// is given back.
func changeFor(total Money, payments []Payment) (Money, error) {
	var cash, other Money
	for _, payment := range payments {
		if payment.Method == PaymentCash {
			cash += payment.Amount
		} else {
			other += payment.Amount
		}
		if cash+other > MaxSaleAmount {
			return 0, ErrAmountOutOfRange
		}
	}
	if other > total {
		return 0, ErrSaleNonCashPaymentExceedsTotal
	}
	if cash+other < total {
		return 0, ErrSalePaymentInsufficient
	}

	return cash + other - total, nil
}

// Replays checks that request is a retry of the request that made this
// sale: with the same idempotency key, a different content is a conflict.
func (sale Sale) Replays(request SaleRequest) error {
	if sale.Fingerprint != request.Fingerprint() {
		return ErrIdempotencyKeyConflict
	}

	return nil
}

// Void returns the sale voided by actor with reason. Only a completed sale
// can be voided, and the reason is required; voiding puts the stock back
// with compensating movements (StockLevel.VoidSale).
func (sale Sale) Void(reason string, actor Actor, now time.Time) (Sale, error) {
	if sale.Status != SaleCompleted {
		return Sale{}, ErrSaleAlreadyVoided
	}
	validReason, err := requiredText(reason, MaxSaleVoidReasonLength, ErrSaleVoidReasonRequired, ErrSaleVoidReasonTooLong)
	if err != nil {
		return Sale{}, err
	}

	voided := sale
	voided.Status = SaleVoided
	voided.VoidedAt = &now
	voided.VoidReason = &validReason
	voidedBy := actor.ID
	voided.VoidedBy = &voidedBy

	return voided, nil
}

// SaleFilter narrows a listing of sales. Zero values don't filter.
type SaleFilter struct {
	// From and To bound CreatedAt: From included, To excluded.
	From       time.Time
	To         time.Time
	CustomerID string
	SellerID   string
	Status     SaleStatus
}

// NewSaleFilter validates a filter: a known status and a range that starts
// before it ends.
func NewSaleFilter(filter SaleFilter) (SaleFilter, error) {
	if filter.Status != "" && !filter.Status.IsValid() {
		return SaleFilter{}, ErrSaleStatusInvalid
	}
	if !filter.From.IsZero() && !filter.To.IsZero() && !filter.From.Before(filter.To) {
		return SaleFilter{}, ErrSaleDateRangeInvalid
	}
	valid := filter
	valid.CustomerID = strings.TrimSpace(filter.CustomerID)
	valid.SellerID = strings.TrimSpace(filter.SellerID)

	return valid, nil
}

// Matches reports whether sale passes the filter.
func (filter SaleFilter) Matches(sale Sale) bool {
	switch {
	case !filter.From.IsZero() && sale.CreatedAt.Before(filter.From):
		return false
	case !filter.To.IsZero() && !sale.CreatedAt.Before(filter.To):
		return false
	case filter.CustomerID != "" && sale.Customer.ID != filter.CustomerID:
		return false
	case filter.SellerID != "" && (sale.Seller == nil || sale.Seller.ID != filter.SellerID):
		return false
	case filter.Status != "" && sale.Status != filter.Status:
		return false
	default:
		return true
	}
}

// Reference is how a stock movement names the sale that caused it.
func (sale Sale) Reference() string {
	return fmt.Sprintf("Venta %d", sale.Number)
}
