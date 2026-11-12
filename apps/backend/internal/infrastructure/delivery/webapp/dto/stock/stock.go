package dto

import (
	"time"

	"market/backend/internal/business/domain"
)

// StockEntryLineRequest is one item of a stock entry. quantity is in
// thousandths of the item's unit and unitCost in cents.
type StockEntryLineRequest struct {
	ItemID   string `json:"itemId"`
	Quantity int64  `json:"quantity"`
	UnitCost *int64 `json:"unitCost"`
}

type StockEntryRequest struct {
	SupplierID *string                 `json:"supplierId"`
	Reference  *string                 `json:"reference"`
	Lines      []StockEntryLineRequest `json:"lines"`
}

func (request StockEntryRequest) Entry() domain.StockEntry {
	lines := make([]domain.StockEntryLine, len(request.Lines))
	for i, line := range request.Lines {
		lines[i] = domain.StockEntryLine{ItemID: line.ItemID, Quantity: domain.Quantity(line.Quantity), UnitCost: moneyPtr(line.UnitCost)}
	}

	return domain.StockEntry{SupplierID: request.SupplierID, Reference: request.Reference, Lines: lines}
}

// StockAdjustmentRequest corrects the stock of an item: quantity is signed,
// in thousandths of the item's unit.
type StockAdjustmentRequest struct {
	ItemID   string `json:"itemId"`
	Quantity int64  `json:"quantity"`
	Reason   string `json:"reason"`
}

func (request StockAdjustmentRequest) Adjustment() domain.StockAdjustment {
	return domain.StockAdjustment{ItemID: request.ItemID, Quantity: domain.Quantity(request.Quantity), Reason: request.Reason}
}

// StockLevelResponse is an item with its stock. updatedAt is null while the
// item has no movements.
type StockLevelResponse struct {
	ItemID     string     `json:"itemId"`
	Code       string     `json:"code"`
	Name       string     `json:"name"`
	CategoryID string     `json:"categoryId"`
	Unit       string     `json:"unit"`
	MinStock   int64      `json:"minStock"`
	Quantity   int64      `json:"quantity"`
	LowStock   bool       `json:"lowStock"`
	UpdatedAt  *time.Time `json:"updatedAt"`
}

func NewStockLevelResponse(itemStock domain.ItemStock) StockLevelResponse {
	var updatedAt *time.Time
	if itemStock.Level.HasMovements() {
		at := itemStock.Level.UpdatedAt
		updatedAt = &at
	}

	return StockLevelResponse{
		ItemID:     itemStock.Item.ID,
		Code:       itemStock.Item.Code,
		Name:       itemStock.Item.Name,
		CategoryID: itemStock.Item.CategoryID,
		Unit:       string(itemStock.Item.Unit),
		MinStock:   int64(itemStock.Item.MinStock),
		Quantity:   int64(itemStock.Level.Quantity),
		LowStock:   itemStock.IsLow(),
		UpdatedAt:  updatedAt,
	}
}

type StockMovementResponse struct {
	ID           string    `json:"id"`
	ItemID       string    `json:"itemId"`
	Type         string    `json:"type"`
	Quantity     int64     `json:"quantity"`
	BalanceAfter int64     `json:"balanceAfter"`
	Reason       *string   `json:"reason"`
	Reference    *string   `json:"reference"`
	SupplierID   *string   `json:"supplierId"`
	UnitCost     *int64    `json:"unitCost"`
	ActorID      string    `json:"actorId"`
	CreatedAt    time.Time `json:"createdAt"`
}

func NewStockMovementResponse(movement domain.StockMovement) StockMovementResponse {
	var unitCost *int64
	if movement.UnitCost != nil {
		cost := int64(*movement.UnitCost)
		unitCost = &cost
	}

	return StockMovementResponse{
		ID:           movement.ID,
		ItemID:       movement.ItemID,
		Type:         string(movement.Type),
		Quantity:     int64(movement.Quantity),
		BalanceAfter: int64(movement.BalanceAfter),
		Reason:       movement.Reason,
		Reference:    movement.Reference,
		SupplierID:   movement.SupplierID,
		UnitCost:     unitCost,
		ActorID:      movement.ActorID,
		CreatedAt:    movement.CreatedAt,
	}
}

// StockEntryResponse holds the movements an entry recorded, one per line in
// the order of the request.
type StockEntryResponse struct {
	Movements []StockMovementResponse `json:"movements"`
}

func NewStockEntryResponse(movements []domain.StockMovement) StockEntryResponse {
	response := StockEntryResponse{Movements: make([]StockMovementResponse, len(movements))}
	for i, movement := range movements {
		response.Movements[i] = NewStockMovementResponse(movement)
	}

	return response
}

func moneyPtr(cents *int64) *domain.Money {
	if cents == nil {
		return nil
	}
	amount := domain.Money(*cents)

	return &amount
}
