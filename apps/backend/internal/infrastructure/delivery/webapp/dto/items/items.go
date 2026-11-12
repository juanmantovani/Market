package dto

import (
	"time"

	"market/backend/internal/business/domain"
)

// ItemRequest is the body of both creating and replacing an item. Amounts
// are integers: salePrice and cost in cents, taxRate in hundredths of a
// percent (2100 = 21 %), minStock in thousandths of the unit.
type ItemRequest struct {
	Code        string   `json:"code"`
	Barcode     *string  `json:"barcode"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	CategoryID  string   `json:"categoryId"`
	SupplierIDs []string `json:"supplierIds"`
	Unit        string   `json:"unit"`
	SalePrice   int64    `json:"salePrice"`
	Cost        *int64   `json:"cost"`
	TaxRate     int64    `json:"taxRate"`
	TracksStock bool     `json:"tracksStock"`
	MinStock    int64    `json:"minStock"`
}

func (request ItemRequest) Fields() domain.ItemFields {
	var cost *domain.Money
	if request.Cost != nil {
		value := domain.Money(*request.Cost)
		cost = &value
	}

	return domain.ItemFields{
		Code:        request.Code,
		Barcode:     request.Barcode,
		Name:        request.Name,
		Description: request.Description,
		CategoryID:  request.CategoryID,
		SupplierIDs: request.SupplierIDs,
		Unit:        domain.Unit(request.Unit),
		SalePrice:   domain.Money(request.SalePrice),
		Cost:        cost,
		TaxRate:     domain.Percent(request.TaxRate),
		TracksStock: request.TracksStock,
		MinStock:    domain.Quantity(request.MinStock),
	}
}

type ItemResponse struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Barcode     *string   `json:"barcode"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CategoryID  string    `json:"categoryId"`
	SupplierIDs []string  `json:"supplierIds"`
	Unit        string    `json:"unit"`
	SalePrice   int64     `json:"salePrice"`
	Cost        *int64    `json:"cost"`
	TaxRate     int64     `json:"taxRate"`
	TracksStock bool      `json:"tracksStock"`
	MinStock    int64     `json:"minStock"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func NewItemResponse(item domain.Item) ItemResponse {
	var cost *int64
	if item.Cost != nil {
		value := int64(*item.Cost)
		cost = &value
	}

	return ItemResponse{
		ID:          item.ID,
		Code:        item.Code,
		Barcode:     item.Barcode,
		Name:        item.Name,
		Description: item.Description,
		CategoryID:  item.CategoryID,
		SupplierIDs: append([]string{}, item.SupplierIDs...),
		Unit:        string(item.Unit),
		SalePrice:   int64(item.SalePrice),
		Cost:        cost,
		TaxRate:     int64(item.TaxRate),
		TracksStock: item.TracksStock,
		MinStock:    int64(item.MinStock),
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}
