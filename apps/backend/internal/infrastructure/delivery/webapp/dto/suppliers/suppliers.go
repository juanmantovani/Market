package dto

import (
	"time"

	"market/backend/internal/business/domain"
)

// SupplierRequest is the body of both creating and replacing a supplier.
type SupplierRequest struct {
	BusinessName string  `json:"businessName"`
	CUIT         *string `json:"cuit"`
	ContactName  *string `json:"contactName"`
	Phone        *string `json:"phone"`
	Email        *string `json:"email"`
	Address      *string `json:"address"`
	Notes        *string `json:"notes"`
}

func (request SupplierRequest) Fields() domain.SupplierFields {
	return domain.SupplierFields{
		BusinessName: request.BusinessName,
		CUIT:         request.CUIT,
		ContactName:  request.ContactName,
		Phone:        request.Phone,
		Email:        request.Email,
		Address:      request.Address,
		Notes:        request.Notes,
	}
}

type SupplierResponse struct {
	ID           string    `json:"id"`
	BusinessName string    `json:"businessName"`
	CUIT         *string   `json:"cuit"`
	ContactName  *string   `json:"contactName"`
	Phone        *string   `json:"phone"`
	Email        *string   `json:"email"`
	Address      *string   `json:"address"`
	Notes        *string   `json:"notes"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func NewSupplierResponse(supplier domain.Supplier) SupplierResponse {
	return SupplierResponse{
		ID:           supplier.ID,
		BusinessName: supplier.BusinessName,
		CUIT:         supplier.CUIT,
		ContactName:  supplier.ContactName,
		Phone:        supplier.Phone,
		Email:        supplier.Email,
		Address:      supplier.Address,
		Notes:        supplier.Notes,
		CreatedAt:    supplier.CreatedAt,
		UpdatedAt:    supplier.UpdatedAt,
	}
}
