package dto

import (
	"time"

	"market/backend/internal/business/usecase/dashboard"
	stockdto "market/backend/internal/infrastructure/delivery/webapp/dto/stock"
)

// MethodTotalResponse is what a payment method collected, in cents.
type MethodTotalResponse struct {
	Method string `json:"method"`
	Amount int64  `json:"amount"`
}

type SalesSummaryResponse struct {
	Count    int                   `json:"count"`
	Total    int64                 `json:"total"`
	ByMethod []MethodTotalResponse `json:"byMethod"`
}

// LowStockResponse lists the first items whose stock reached their minimum;
// total counts all of them.
type LowStockResponse struct {
	Total int                           `json:"total"`
	Items []stockdto.StockLevelResponse `json:"items"`
}

// SummaryResponse is the summary of a period: from included, to excluded.
// The sales don't count voided ones, and cash is net of the change.
type SummaryResponse struct {
	From     time.Time            `json:"from"`
	To       time.Time            `json:"to"`
	Sales    SalesSummaryResponse `json:"sales"`
	LowStock LowStockResponse     `json:"lowStock"`
}

func NewSummaryResponse(summary dashboard.Summary) SummaryResponse {
	byMethod := make([]MethodTotalResponse, len(summary.Sales.ByMethod))
	for i, total := range summary.Sales.ByMethod {
		byMethod[i] = MethodTotalResponse{Method: string(total.Method), Amount: int64(total.Amount)}
	}
	items := make([]stockdto.StockLevelResponse, len(summary.LowStock))
	for i, itemStock := range summary.LowStock {
		items[i] = stockdto.NewStockLevelResponse(itemStock)
	}

	return SummaryResponse{
		From: summary.Period.From,
		To:   summary.Period.To,
		Sales: SalesSummaryResponse{
			Count:    summary.Sales.Count,
			Total:    int64(summary.Sales.Total),
			ByMethod: byMethod,
		},
		LowStock: LowStockResponse{Total: summary.LowStockTotal, Items: items},
	}
}
