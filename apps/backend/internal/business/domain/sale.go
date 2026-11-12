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
	ErrSaleDiscountInvalid            = &Error{Kind: KindInvalid, Code: "sale_discount_invalid", Message: "El descuento va de 0 a 100 %"}
	ErrSaleCustomerNotFound           = &Error{Kind: KindInvalid, Code: "sale_customer_not_found", Message: "El cliente elegido no existe"}
	ErrSaleSellerNotFound             = &Error{Kind: KindInvalid, Code: "sale_seller_not_found", Message: "El vendedor elegido no existe"}
	ErrSaleTotalMismatch              = &Error{Kind: KindConflict, Code: "sale_total_mismatch", Message: "El total cambió mientras se armaba la venta: revisar el carrito antes de cobrar"}
	ErrSaleTooManyPayments            = &Error{Kind: KindInvalid, Code: "sale_too_many_payments", Message: "La venta admite hasta 10 pagos"}
	ErrPaymentMethodInvalid           = &Error{Kind: KindInvalid, Code: "payment_method_invalid", Message: "El medio de pago no es válido"}
	ErrPaymentAmountInvalid           = &Error{Kind: KindInvalid, Code: "payment_amount_invalid", Message: "El importe de cada pago tiene que ser mayor a cero"}
	ErrPaymentReferenceTooLong        = &Error{Kind: KindInvalid, Code: "payment_reference_too_long", Message: "La referencia del pago admite hasta 60 caracteres"}
	ErrSalePaymentInsufficient        = &Error{Kind: KindInvalid, Code: "sale_payment_insufficient", Message: "Los pagos no cubren el total de la venta"}
	ErrSaleNonCashPaymentExceedsTotal = &Error{Kind: KindInvalid, Code: "sale_non_cash_payment_exceeds_total", Message: "Solo el efectivo puede superar el total: ajustar los otros medios de pago"}
)

// SaleRequestLine is one item the cashier wants to sell. DiscountPercent
// applies to this line only; the "general discount" of the counter is the
// same percent on every line.
type SaleRequestLine struct {
	ItemID          string
	Quantity        Quantity
	DiscountPercent Percent
}

// SaleRequest is what the counter asks to sell, before reading any stored
// data. ExpectedTotal is the total the cashier saw: the sale is refused if
// the server computes another one.
type SaleRequest struct {
	IdempotencyKey string
	// CustomerID is optional: a sale without customer goes to Consumidor
	// Final.
	CustomerID    *string
	SellerID      *string
	Lines         []SaleRequestLine
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
	payments, err := validatePayments(request.Payments)
	if err != nil {
		return SaleRequest{}, err
	}

	return SaleRequest{
		IdempotencyKey: key,
		CustomerID:     optionalID(request.CustomerID),
		SellerID:       optionalID(request.SellerID),
		Lines:          lines,
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
		if !line.DiscountPercent.IsValid() {
			return nil, ErrSaleDiscountInvalid
		}
		valid = append(valid, SaleRequestLine{ItemID: itemID, Quantity: line.Quantity, DiscountPercent: line.DiscountPercent})
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
	fmt.Fprintf(hash, "customer=%q seller=%q expected=%d\n", valueOr(request.CustomerID), valueOr(request.SellerID), request.ExpectedTotal)
	for _, line := range request.Lines {
		fmt.Fprintf(hash, "line=%q %d %d\n", line.ItemID, line.Quantity, line.DiscountPercent)
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
	UnitPrice       Money
	DiscountPercent Percent
	DiscountAmount  Money
	// LineTotal is UnitPrice × Quantity minus DiscountAmount.
	LineTotal Money
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
	Seller        *SaleSeller
	Lines         []SaleLine
	Subtotal      Money
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
		sale.Subtotal += line.LineTotal + line.DiscountAmount
		sale.DiscountTotal += line.DiscountAmount
		if sale.Subtotal > MaxSaleAmount {
			return Sale{}, ErrAmountOutOfRange
		}
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

// newSaleLine prices one line: the amount is rounded to the cent, then the
// discount is rounded to the cent and taken from it, so the lines of a
// ticket always add up to its total.
func newSaleLine(item Item, requested SaleRequestLine) (SaleLine, error) {
	if !item.Unit.Accepts(requested.Quantity) {
		return SaleLine{}, ErrSaleQuantityNotWhole
	}
	amount, err := item.SalePrice.Times(requested.Quantity)
	if err != nil {
		return SaleLine{}, err
	}
	discount, err := amount.Percent(requested.DiscountPercent)
	if err != nil {
		return SaleLine{}, err
	}

	return SaleLine{
		ItemID:          item.ID,
		Code:            item.Code,
		Name:            item.Name,
		Unit:            item.Unit,
		TaxRate:         item.TaxRate,
		Quantity:        requested.Quantity,
		UnitPrice:       item.SalePrice,
		DiscountPercent: requested.DiscountPercent,
		DiscountAmount:  discount,
		LineTotal:       amount - discount,
	}, nil
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

// Reference is how a stock movement names the sale that caused it.
func (sale Sale) Reference() string {
	return fmt.Sprintf("Venta %d", sale.Number)
}
