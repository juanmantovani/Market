package dto

import (
	"time"

	"market/backend/internal/business/domain"
)

// SellerRequest is the body of both creating and replacing a seller. The
// commission travels in hundredths of a percent (2.5 % is 250).
type SellerRequest struct {
	Name           string `json:"name"`
	Code           string `json:"code"`
	CommissionRate *int64 `json:"commissionRate"`
}

func (request SellerRequest) Fields() domain.SellerFields {
	var rate *domain.Percent
	if request.CommissionRate != nil {
		value := domain.Percent(*request.CommissionRate)
		rate = &value
	}

	return domain.SellerFields{Name: request.Name, Code: request.Code, CommissionRate: rate}
}

type SellerResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Code           string    `json:"code"`
	CommissionRate *int64    `json:"commissionRate"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func NewSellerResponse(seller domain.Seller) SellerResponse {
	var rate *int64
	if seller.CommissionRate != nil {
		value := int64(*seller.CommissionRate)
		rate = &value
	}

	return SellerResponse{
		ID:             seller.ID,
		Name:           seller.Name,
		Code:           seller.Code,
		CommissionRate: rate,
		CreatedAt:      seller.CreatedAt,
		UpdatedAt:      seller.UpdatedAt,
	}
}
