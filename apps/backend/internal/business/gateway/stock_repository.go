package gateway

import (
	"context"

	"market/backend/internal/business/domain"
)

type StockLevelFilter struct {
	// Search matches the name, code or barcode of the item, ignoring case
	// and accents.
	Search string
	// CategoryID, when set, keeps only the items of that category.
	CategoryID string
	// LowStock keeps only the items whose stock reached their minimum, as
	// domain.IsLowStock decides.
	LowStock bool
	Page     domain.PageRequest
}

type StockMovementFilter struct {
	// ItemID, when set, keeps only the movements of that item.
	ItemID string
	// Type, when set, keeps only the movements of that type.
	Type domain.MovementType
	Page domain.PageRequest
}

// StockRepository stores stock levels and movements. Every implementation
// must pass the contract in infrastructure/repository/repositorytest.
//
// A use case reads a level, lets the domain compute the movement and records
// it within the same TxManager unit of work, so no concurrent movement of the
// same item slips in between.
type StockRepository interface {
	// GetLevel returns the stock level of an item, locking it until the unit
	// of work ends. An item without movements has a zero level.
	GetLevel(ctx context.Context, itemID string) (domain.StockLevel, error)
	// Record stores a movement, assigning its ID, and leaves the level of its
	// item at BalanceAfter as of CreatedAt. It returns ErrStockItemNotFound
	// when the item doesn't exist and ErrStockSupplierNotFound when the
	// movement names a supplier that doesn't.
	Record(ctx context.Context, movement domain.StockMovement) (domain.StockMovement, error)
	// ListLevels returns one page of items with their stock level, sorted by
	// item name. Every item is listed, with a zero level if it has no
	// movements.
	ListLevels(ctx context.Context, filter StockLevelFilter) (domain.Page[domain.ItemStock], error)
	// ListMovements returns one page of movements, newest first; movements
	// recorded at the same instant come in reverse order of recording.
	ListMovements(ctx context.Context, filter StockMovementFilter) (domain.Page[domain.StockMovement], error)
}
