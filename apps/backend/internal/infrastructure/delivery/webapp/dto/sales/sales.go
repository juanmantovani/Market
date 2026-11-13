package dto

import (
	"time"

	"market/backend/internal/business/domain"
)

// Discount is taken off a line or off the whole sale: type "percent" with
// value in hundredths of a percent (10 % is 1000), or type "amount" with
// value in cents. null means no discount.
type Discount struct {
	Type  string `json:"type"`
	Value int64  `json:"value"`
}

func (discount *Discount) domain() *domain.Discount {
	if discount == nil {
		return nil
	}

	return &domain.Discount{Type: domain.DiscountType(discount.Type), Value: discount.Value}
}

func newDiscount(discount *domain.Discount) *Discount {
	if discount == nil {
		return nil
	}

	return &Discount{Type: string(discount.Type), Value: discount.Value}
}

// SaleLineRequest is one item to sell: quantity in thousandths of its unit
// and an optional discount for this line.
type SaleLineRequest struct {
	ItemID   string    `json:"itemId"`
	Quantity int64     `json:"quantity"`
	Discount *Discount `json:"discount"`
}

// PaymentRequest is one payment, in cents.
type PaymentRequest struct {
	Method    string  `json:"method"`
	Amount    int64   `json:"amount"`
	Reference *string `json:"reference"`
}

// CreateSaleRequest is the body of POST /api/v1/sales; the idempotency key
// travels in the Idempotency-Key header. cartDiscount is the optional
// general discount of the whole sale, and expectedTotal the total the
// cashier saw, in cents.
type CreateSaleRequest struct {
	CustomerID    *string           `json:"customerId"`
	SellerID      *string           `json:"sellerId"`
	Lines         []SaleLineRequest `json:"lines"`
	CartDiscount  *Discount         `json:"cartDiscount"`
	Payments      []PaymentRequest  `json:"payments"`
	ExpectedTotal int64             `json:"expectedTotal"`
}

func (request CreateSaleRequest) Sale(idempotencyKey string) domain.SaleRequest {
	lines := make([]domain.SaleRequestLine, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = domain.SaleRequestLine{
			ItemID:   line.ItemID,
			Quantity: domain.Quantity(line.Quantity),
			Discount: line.Discount.domain(),
		}
	}
	payments := make([]domain.Payment, len(request.Payments))
	for i, payment := range request.Payments {
		payments[i] = domain.Payment{Method: domain.PaymentMethod(payment.Method), Amount: domain.Money(payment.Amount), Reference: payment.Reference}
	}

	return domain.SaleRequest{
		IdempotencyKey: idempotencyKey,
		CustomerID:     request.CustomerID,
		SellerID:       request.SellerID,
		Lines:          lines,
		CartDiscount:   request.CartDiscount.domain(),
		Payments:       payments,
		ExpectedTotal:  domain.Money(request.ExpectedTotal),
	}
}

// SaleCustomerResponse is the customer as it was when the sale was made.
type SaleCustomerResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	DocumentType   string `json:"documentType"`
	DocumentNumber string `json:"documentNumber"`
	TaxCondition   string `json:"taxCondition"`
}

// SaleSellerResponse is the seller as it was when the sale was made.
type SaleSellerResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// SaleLineResponse is one item sold, with the code, name, unit, tax rate
// and price it had when it was sold. discountAmount is what the line's own
// discount took off, cartDiscountAmount its share of the general discount,
// and lineTotal what is left.
type SaleLineResponse struct {
	ItemID             string    `json:"itemId"`
	Code               string    `json:"code"`
	Name               string    `json:"name"`
	Unit               string    `json:"unit"`
	TaxRate            int64     `json:"taxRate"`
	Quantity           int64     `json:"quantity"`
	UnitPrice          int64     `json:"unitPrice"`
	Discount           *Discount `json:"discount"`
	DiscountAmount     int64     `json:"discountAmount"`
	CartDiscountAmount int64     `json:"cartDiscountAmount"`
	LineTotal          int64     `json:"lineTotal"`
}

type PaymentResponse struct {
	Method    string  `json:"method"`
	Amount    int64   `json:"amount"`
	Reference *string `json:"reference"`
}

// SaleResponse is a sale with every amount in cents. seller is null for a
// sale nobody in particular attended; the void fields are null while the
// sale is completed. discountTotal adds the line discounts and
// cartDiscountAmount, what the general discount took off.
type SaleResponse struct {
	ID                 string               `json:"id"`
	Number             int64                `json:"number"`
	Status             string               `json:"status"`
	Customer           SaleCustomerResponse `json:"customer"`
	Seller             *SaleSellerResponse  `json:"seller"`
	Lines              []SaleLineResponse   `json:"lines"`
	Subtotal           int64                `json:"subtotal"`
	CartDiscount       *Discount            `json:"cartDiscount"`
	CartDiscountAmount int64                `json:"cartDiscountAmount"`
	DiscountTotal      int64                `json:"discountTotal"`
	Total              int64                `json:"total"`
	Payments           []PaymentResponse    `json:"payments"`
	Change             int64                `json:"change"`
	ActorID            string               `json:"actorId"`
	CreatedAt          time.Time            `json:"createdAt"`
	VoidedAt           *time.Time           `json:"voidedAt"`
	VoidReason         *string              `json:"voidReason"`
	VoidedBy           *string              `json:"voidedBy"`
}

func NewSaleResponse(sale domain.Sale) SaleResponse {
	response := SaleResponse{
		ID:     sale.ID,
		Number: sale.Number,
		Status: string(sale.Status),
		Customer: SaleCustomerResponse{
			ID:             sale.Customer.ID,
			Name:           sale.Customer.Name,
			DocumentType:   string(sale.Customer.DocumentType),
			DocumentNumber: sale.Customer.DocumentNumber,
			TaxCondition:   string(sale.Customer.TaxCondition),
		},
		Lines:              make([]SaleLineResponse, len(sale.Lines)),
		Subtotal:           int64(sale.Subtotal),
		CartDiscount:       newDiscount(sale.CartDiscount),
		CartDiscountAmount: int64(sale.CartDiscountAmount),
		DiscountTotal:      int64(sale.DiscountTotal),
		Total:              int64(sale.Total),
		Payments:           make([]PaymentResponse, len(sale.Payments)),
		Change:             int64(sale.Change),
		ActorID:            sale.ActorID,
		CreatedAt:          sale.CreatedAt,
		VoidedAt:           sale.VoidedAt,
		VoidReason:         sale.VoidReason,
		VoidedBy:           sale.VoidedBy,
	}
	if sale.Seller != nil {
		response.Seller = &SaleSellerResponse{ID: sale.Seller.ID, Name: sale.Seller.Name, Code: sale.Seller.Code}
	}
	for i, line := range sale.Lines {
		response.Lines[i] = SaleLineResponse{
			ItemID:             line.ItemID,
			Code:               line.Code,
			Name:               line.Name,
			Unit:               string(line.Unit),
			TaxRate:            int64(line.TaxRate),
			Quantity:           int64(line.Quantity),
			UnitPrice:          int64(line.UnitPrice),
			Discount:           newDiscount(line.Discount),
			DiscountAmount:     int64(line.DiscountAmount),
			CartDiscountAmount: int64(line.CartDiscountAmount),
			LineTotal:          int64(line.LineTotal),
		}
	}
	for i, payment := range sale.Payments {
		response.Payments[i] = PaymentResponse{Method: string(payment.Method), Amount: int64(payment.Amount), Reference: payment.Reference}
	}

	return response
}

// VoidSaleRequest is the body of POST /api/v1/sales/{id}/void.
type VoidSaleRequest struct {
	Reason string `json:"reason"`
}
