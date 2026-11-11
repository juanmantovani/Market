package dto

import (
	"time"

	"market/backend/internal/business/domain"
)

// CustomerRequest is the body of both creating and replacing a customer.
type CustomerRequest struct {
	Name           string  `json:"name"`
	DocumentType   string  `json:"documentType"`
	DocumentNumber string  `json:"documentNumber"`
	TaxCondition   string  `json:"taxCondition"`
	Email          *string `json:"email"`
	Phone          *string `json:"phone"`
	Address        *string `json:"address"`
	Notes          *string `json:"notes"`
}

func (request CustomerRequest) Fields() domain.CustomerFields {
	return domain.CustomerFields{
		Name:           request.Name,
		DocumentType:   domain.DocumentType(request.DocumentType),
		DocumentNumber: request.DocumentNumber,
		TaxCondition:   domain.TaxCondition(request.TaxCondition),
		Email:          request.Email,
		Phone:          request.Phone,
		Address:        request.Address,
		Notes:          request.Notes,
	}
}

type CustomerResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	DocumentType   string    `json:"documentType"`
	DocumentNumber string    `json:"documentNumber"`
	TaxCondition   string    `json:"taxCondition"`
	Email          *string   `json:"email"`
	Phone          *string   `json:"phone"`
	Address        *string   `json:"address"`
	Notes          *string   `json:"notes"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func NewCustomerResponse(customer domain.Customer) CustomerResponse {
	return CustomerResponse{
		ID:             customer.ID,
		Name:           customer.Name,
		DocumentType:   string(customer.DocumentType),
		DocumentNumber: customer.DocumentNumber,
		TaxCondition:   string(customer.TaxCondition),
		Email:          customer.Email,
		Phone:          customer.Phone,
		Address:        customer.Address,
		Notes:          customer.Notes,
		CreatedAt:      customer.CreatedAt,
		UpdatedAt:      customer.UpdatedAt,
	}
}
