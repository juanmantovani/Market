package gatewayfake

import (
	"context"
	"fmt"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

// StockRepository is a fake gateway.StockRepository for use case tests. It
// records every call and returns the configured results. Levels holds the
// level of each item; Record moves it to the movement's balance, so later
// reads see what was recorded.
type StockRepository struct {
	Levels      map[string]domain.StockLevel
	GetLevelErr error
	LevelIDs    []string

	RecordErr error
	Recorded  []domain.StockMovement

	ListLevelsResult  domain.Page[domain.ItemStock]
	ListLevelsErr     error
	ListLevelsFilters []gateway.StockLevelFilter

	ListMovementsResult  domain.Page[domain.StockMovement]
	ListMovementsErr     error
	ListMovementsFilters []gateway.StockMovementFilter
}

func (fake *StockRepository) GetLevel(_ context.Context, itemID string) (domain.StockLevel, error) {
	fake.LevelIDs = append(fake.LevelIDs, itemID)
	if fake.GetLevelErr != nil {
		return domain.StockLevel{}, fake.GetLevelErr
	}
	if level, ok := fake.Levels[itemID]; ok {
		return level, nil
	}

	return domain.StockLevel{ItemID: itemID}, nil
}

func (fake *StockRepository) Record(_ context.Context, movement domain.StockMovement) (domain.StockMovement, error) {
	if fake.RecordErr != nil {
		return domain.StockMovement{}, fake.RecordErr
	}
	movement.ID = fmt.Sprintf("movement-%d", len(fake.Recorded)+1)
	fake.Recorded = append(fake.Recorded, movement)
	if fake.Levels == nil {
		fake.Levels = map[string]domain.StockLevel{}
	}
	fake.Levels[movement.ItemID] = domain.StockLevel{ItemID: movement.ItemID, Quantity: movement.BalanceAfter, UpdatedAt: movement.CreatedAt}

	return movement, nil
}

func (fake *StockRepository) ListLevels(_ context.Context, filter gateway.StockLevelFilter) (domain.Page[domain.ItemStock], error) {
	fake.ListLevelsFilters = append(fake.ListLevelsFilters, filter)
	if fake.ListLevelsErr != nil {
		return domain.Page[domain.ItemStock]{}, fake.ListLevelsErr
	}

	return fake.ListLevelsResult, nil
}

func (fake *StockRepository) ListMovements(_ context.Context, filter gateway.StockMovementFilter) (domain.Page[domain.StockMovement], error) {
	fake.ListMovementsFilters = append(fake.ListMovementsFilters, filter)
	if fake.ListMovementsErr != nil {
		return domain.Page[domain.StockMovement]{}, fake.ListMovementsErr
	}

	return fake.ListMovementsResult, nil
}
