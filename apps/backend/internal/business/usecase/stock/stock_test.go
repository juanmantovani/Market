package stock

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
	"market/backend/internal/business/gateway/gatewayfake"
)

var (
	admin   = domain.Actor{ID: "u-admin", Role: domain.RoleAdmin}
	manager = domain.Actor{ID: "u-manager", Role: domain.RoleManager}
	cashier = domain.Actor{ID: "u-cashier", Role: domain.RoleCashier}

	errDown = errors.New("connection refused")
)

func ptr[T any](value T) *T {
	return &value
}

func item(id string, unit domain.Unit, cost *domain.Money) domain.Item {
	return domain.Item{ID: id, ItemFields: domain.ItemFields{
		Code: id, Name: "Artículo " + id, CategoryID: "category-1", Unit: unit, SalePrice: 100_000, TaxRate: 2100, Cost: cost,
	}}
}

func assertError(t *testing.T, err error, want *domain.Error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %s", err, want.Code)
	}
}

type entryFakes struct {
	tx        *gatewayfake.TxManager
	stock     *gatewayfake.StockRepository
	items     *gatewayfake.ItemRepository
	suppliers *gatewayfake.SupplierRepository
}

func newEntryFakes() entryFakes {
	return entryFakes{
		tx:    &gatewayfake.TxManager{},
		stock: &gatewayfake.StockRepository{Levels: map[string]domain.StockLevel{"yerba": {ItemID: "yerba", Quantity: 4000, UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}}},
		items: &gatewayfake.ItemRepository{Stored: map[string]domain.Item{
			"yerba": item("yerba", domain.UnitPiece, ptr(domain.Money(310_000))),
			"queso": item("queso", domain.UnitKilogram, nil),
		}},
		suppliers: &gatewayfake.SupplierRepository{GetResult: domain.Supplier{ID: "supplier-1"}},
	}
}

func (fakes entryFakes) useCase() RegisterStockEntry {
	return NewRegisterStockEntry(fakes.tx, fakes.stock, fakes.items, fakes.suppliers)
}

func entry() domain.StockEntry {
	return domain.StockEntry{
		SupplierID: ptr("supplier-1"),
		Reference:  ptr("Remito 12"),
		Lines: []domain.StockEntryLine{
			{ItemID: "yerba", Quantity: domain.WholeQuantity(12), UnitCost: ptr(domain.Money(320_000))},
			{ItemID: "queso", Quantity: 2500},
		},
	}
}

func TestRegisterStockEntry(t *testing.T) {
	fakes := newEntryFakes()
	before := time.Now().UTC()

	movements, err := fakes.useCase().Execute(context.Background(), RegisterStockEntryInput{Actor: manager, Entry: entry()})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if fakes.tx.Calls != 1 || len(movements) != 2 || len(fakes.stock.Recorded) != 2 {
		t.Fatalf("tx calls %d, movements %+v", fakes.tx.Calls, movements)
	}
	yerba, queso := movements[0], movements[1]
	if yerba.ID == "" || yerba.ItemID != "yerba" || yerba.Type != domain.MovementEntry || yerba.Quantity != 12_000 || yerba.BalanceAfter != 16_000 ||
		*yerba.SupplierID != "supplier-1" || *yerba.Reference != "Remito 12" || *yerba.UnitCost != 320_000 ||
		yerba.ActorID != "u-manager" || yerba.CreatedAt.Before(before) || yerba.CreatedAt.Location() != time.UTC {
		t.Fatalf("yerba movement = %+v", yerba)
	}
	if queso.ItemID != "queso" || queso.BalanceAfter != 2500 || queso.UnitCost != nil || !queso.CreatedAt.Equal(yerba.CreatedAt) {
		t.Fatalf("queso movement = %+v", queso)
	}
	if len(fakes.suppliers.GetIDs) != 1 || fakes.suppliers.GetIDs[0] != "supplier-1" {
		t.Fatalf("supplier lookups = %v", fakes.suppliers.GetIDs)
	}
	if len(fakes.items.Updated) != 1 || fakes.items.Updated[0].ID != "yerba" || *fakes.items.Updated[0].Cost != 320_000 {
		t.Fatalf("items updated = %+v, want only yerba with its new cost", fakes.items.Updated)
	}
}

func TestRegisterStockEntryWithoutSupplierOrNewCost(t *testing.T) {
	fakes := newEntryFakes()
	sameCost := domain.StockEntry{Lines: []domain.StockEntryLine{{ItemID: "yerba", Quantity: 1000, UnitCost: ptr(domain.Money(310_000))}}}

	movements, err := fakes.useCase().Execute(context.Background(), RegisterStockEntryInput{Actor: admin, Entry: sameCost})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(movements) != 1 || movements[0].SupplierID != nil || len(fakes.suppliers.GetIDs) != 0 || len(fakes.items.Updated) != 0 {
		t.Fatalf("movements %+v, supplier lookups %v, item updates %d", movements, fakes.suppliers.GetIDs, len(fakes.items.Updated))
	}
}

func TestRegisterStockEntryRejectsBeforeTheUnitOfWork(t *testing.T) {
	tests := []struct {
		name  string
		input RegisterStockEntryInput
		want  *domain.Error
	}{
		{"cashier", RegisterStockEntryInput{Actor: cashier, Entry: entry()}, domain.ErrForbidden},
		{"no lines", RegisterStockEntryInput{Actor: admin, Entry: domain.StockEntry{}}, domain.ErrStockEntryLinesRequired},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakes := newEntryFakes()

			_, err := fakes.useCase().Execute(context.Background(), test.input)

			assertError(t, err, test.want)
			if fakes.tx.Calls != 0 {
				t.Fatal("the unit of work started")
			}
		})
	}
}

func TestRegisterStockEntryErrors(t *testing.T) {
	tests := []struct {
		name    string
		arrange func(fakes *entryFakes)
		entry   func() domain.StockEntry
		want    *domain.Error
	}{
		{"unit of work fails", func(fakes *entryFakes) { fakes.tx.Err = errDown }, entry, domain.ErrStorageUnavailable},
		{"unknown supplier", func(fakes *entryFakes) { fakes.suppliers.GetErr = domain.ErrSupplierNotFound }, entry, domain.ErrStockSupplierNotFound},
		{"supplier storage down", func(fakes *entryFakes) { fakes.suppliers.GetErr = errDown }, entry, domain.ErrStorageUnavailable},
		{"unknown item", func(fakes *entryFakes) { delete(fakes.items.Stored, "queso") }, entry, domain.ErrStockItemNotFound},
		{"level unavailable", func(fakes *entryFakes) { fakes.stock.GetLevelErr = errDown }, entry, domain.ErrStorageUnavailable},
		{"fraction of a unit", func(*entryFakes) {}, func() domain.StockEntry {
			return domain.StockEntry{Lines: []domain.StockEntryLine{{ItemID: "yerba", Quantity: 500}}}
		}, domain.ErrStockQuantityNotWhole},
		{"record fails", func(fakes *entryFakes) { fakes.stock.RecordErr = errDown }, entry, domain.ErrStorageUnavailable},
		{"stored item is invalid", func(fakes *entryFakes) { fakes.items.Stored["yerba"] = item("yerba", "box", nil) }, func() domain.StockEntry {
			return domain.StockEntry{Lines: []domain.StockEntryLine{{ItemID: "yerba", Quantity: 1000, UnitCost: ptr(domain.Money(1))}}}
		}, domain.ErrItemUnitInvalid},
		{"cost update fails", func(fakes *entryFakes) { fakes.items.UpdateErr = errDown }, entry, domain.ErrStorageUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakes := newEntryFakes()
			test.arrange(&fakes)

			movements, err := fakes.useCase().Execute(context.Background(), RegisterStockEntryInput{Actor: admin, Entry: test.entry()})

			assertError(t, err, test.want)
			if movements != nil {
				t.Fatalf("movements = %+v, want none", movements)
			}
		})
	}
}

func adjustmentFakes() (*gatewayfake.TxManager, *gatewayfake.StockRepository, *gatewayfake.ItemRepository) {
	return &gatewayfake.TxManager{},
		&gatewayfake.StockRepository{Levels: map[string]domain.StockLevel{"queso": {ItemID: "queso", Quantity: 1800, UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}}},
		&gatewayfake.ItemRepository{Stored: map[string]domain.Item{"queso": item("queso", domain.UnitKilogram, nil)}}
}

func TestRegisterStockAdjustment(t *testing.T) {
	tx, stock, items := adjustmentFakes()

	movement, err := NewRegisterStockAdjustment(tx, stock, items).Execute(context.Background(), RegisterStockAdjustmentInput{
		Actor:      manager,
		Adjustment: domain.StockAdjustment{ItemID: " queso ", Quantity: -300, Reason: " Vencido "},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if tx.Calls != 1 || movement.ID == "" || movement.Type != domain.MovementAdjustment || movement.Quantity != -300 ||
		movement.BalanceAfter != 1500 || *movement.Reason != "Vencido" || movement.ActorID != "u-manager" {
		t.Fatalf("movement = %+v (tx calls %d)", movement, tx.Calls)
	}
	if stock.Levels["queso"].Quantity != 1500 {
		t.Fatalf("level = %+v", stock.Levels["queso"])
	}
}

func TestRegisterStockAdjustmentErrors(t *testing.T) {
	valid := domain.StockAdjustment{ItemID: "queso", Quantity: -300, Reason: "Vencido"}

	tests := []struct {
		name       string
		actor      domain.Actor
		adjustment domain.StockAdjustment
		arrange    func(*gatewayfake.TxManager, *gatewayfake.StockRepository, *gatewayfake.ItemRepository)
		want       *domain.Error
	}{
		{"cashier", cashier, valid, nil, domain.ErrForbidden},
		{"no reason", admin, domain.StockAdjustment{ItemID: "queso", Quantity: 1000}, nil, domain.ErrStockReasonRequired},
		{"unit of work fails", admin, valid, func(tx *gatewayfake.TxManager, _ *gatewayfake.StockRepository, _ *gatewayfake.ItemRepository) {
			tx.Err = errDown
		}, domain.ErrStorageUnavailable},
		{"unknown item", admin, domain.StockAdjustment{ItemID: "missing", Quantity: 1000, Reason: "Conteo"}, nil, domain.ErrStockItemNotFound},
		{"level unavailable", admin, valid, func(_ *gatewayfake.TxManager, stock *gatewayfake.StockRepository, _ *gatewayfake.ItemRepository) {
			stock.GetLevelErr = errDown
		}, domain.ErrStorageUnavailable},
		{"below zero", admin, domain.StockAdjustment{ItemID: "queso", Quantity: -2000, Reason: "Conteo"}, nil, domain.ErrInsufficientStock},
		{"record fails", admin, valid, func(_ *gatewayfake.TxManager, stock *gatewayfake.StockRepository, _ *gatewayfake.ItemRepository) {
			stock.RecordErr = errDown
		}, domain.ErrStorageUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tx, stock, items := adjustmentFakes()
			if test.arrange != nil {
				test.arrange(tx, stock, items)
			}

			_, err := NewRegisterStockAdjustment(tx, stock, items).Execute(context.Background(), RegisterStockAdjustmentInput{Actor: test.actor, Adjustment: test.adjustment})

			assertError(t, err, test.want)
			if len(stock.Recorded) != 0 {
				t.Fatalf("recorded %+v", stock.Recorded)
			}
		})
	}
}

func TestRegisterInitialStock(t *testing.T) {
	tx := &gatewayfake.TxManager{}
	stock := &gatewayfake.StockRepository{}
	items := &gatewayfake.ItemRepository{Stored: map[string]domain.Item{"yerba": item("yerba", domain.UnitPiece, nil)}}

	movement, err := NewRegisterInitialStock(tx, stock, items).Execute(context.Background(), RegisterInitialStockInput{
		Actor: admin, ItemID: "yerba", Quantity: domain.WholeQuantity(4),
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if tx.Calls != 1 || movement.Type != domain.MovementInitial || movement.BalanceAfter != 4000 || movement.ID == "" {
		t.Fatalf("movement = %+v", movement)
	}

	_, err = NewRegisterInitialStock(tx, stock, items).Execute(context.Background(), RegisterInitialStockInput{
		Actor: admin, ItemID: "yerba", Quantity: domain.WholeQuantity(4),
	})
	assertError(t, err, domain.ErrStockAlreadyInitialized)
}

func TestRegisterInitialStockErrors(t *testing.T) {
	stored := map[string]domain.Item{"yerba": item("yerba", domain.UnitPiece, nil)}
	valid := RegisterInitialStockInput{Actor: admin, ItemID: "yerba", Quantity: 1000}

	tests := []struct {
		name  string
		tx    *gatewayfake.TxManager
		stock *gatewayfake.StockRepository
		input RegisterInitialStockInput
		want  *domain.Error
	}{
		{"cashier", &gatewayfake.TxManager{}, &gatewayfake.StockRepository{}, RegisterInitialStockInput{Actor: cashier, ItemID: "yerba", Quantity: 1000}, domain.ErrForbidden},
		{"unit of work fails", &gatewayfake.TxManager{Err: errDown}, &gatewayfake.StockRepository{}, valid, domain.ErrStorageUnavailable},
		{"unknown item", &gatewayfake.TxManager{}, &gatewayfake.StockRepository{}, RegisterInitialStockInput{Actor: admin, ItemID: "missing", Quantity: 1000}, domain.ErrStockItemNotFound},
		{"level unavailable", &gatewayfake.TxManager{}, &gatewayfake.StockRepository{GetLevelErr: errDown}, valid, domain.ErrStorageUnavailable},
		{"zero", &gatewayfake.TxManager{}, &gatewayfake.StockRepository{}, RegisterInitialStockInput{Actor: admin, ItemID: "yerba"}, domain.ErrStockQuantityInvalid},
		{"record fails", &gatewayfake.TxManager{}, &gatewayfake.StockRepository{RecordErr: errDown}, valid, domain.ErrStorageUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			items := &gatewayfake.ItemRepository{Stored: stored}

			_, err := NewRegisterInitialStock(test.tx, test.stock, items).Execute(context.Background(), test.input)

			assertError(t, err, test.want)
		})
	}
}

func TestListStockLevels(t *testing.T) {
	page := domain.Page[domain.ItemStock]{Items: []domain.ItemStock{{Item: item("yerba", domain.UnitPiece, nil)}}, Total: 1, Page: 2, Limit: 10}
	repository := &gatewayfake.StockRepository{ListLevelsResult: page}

	got, err := NewListStockLevels(repository).Execute(context.Background(), ListStockLevelsInput{
		Actor: cashier, Search: " yer ", CategoryID: " category-1 ", ItemIDs: []string{" item-1", "", "item-2 "}, LowStock: true, Page: 2, Limit: 10,
	})
	if err != nil || got.Total != 1 {
		t.Fatalf("Execute() = %+v, %v", got, err)
	}
	want := gateway.StockLevelFilter{
		Search: "yer", CategoryID: "category-1", ItemIDs: []string{"item-1", "item-2"}, LowStock: true, Page: domain.PageRequest{Page: 2, Limit: 10},
	}
	if !reflect.DeepEqual(repository.ListLevelsFilters[0], want) {
		t.Fatalf("filter = %+v, want %+v", repository.ListLevelsFilters[0], want)
	}
}

func TestListStockLevelsWithoutItemIDs(t *testing.T) {
	for _, itemIDs := range [][]string{nil, {}, {"", "  "}} {
		repository := &gatewayfake.StockRepository{}

		if _, err := NewListStockLevels(repository).Execute(context.Background(), ListStockLevelsInput{Actor: cashier, ItemIDs: itemIDs}); err != nil {
			t.Fatalf("Execute(%q) error = %v", itemIDs, err)
		}
		if got := repository.ListLevelsFilters[0].ItemIDs; got != nil {
			t.Fatalf("ItemIDs for %q = %q, want nil so that no filter applies", itemIDs, got)
		}
	}
}

func TestListStockLevelsErrors(t *testing.T) {
	_, err := NewListStockLevels(&gatewayfake.StockRepository{}).Execute(context.Background(), ListStockLevelsInput{})
	assertError(t, err, domain.ErrForbidden)

	_, err = NewListStockLevels(&gatewayfake.StockRepository{}).Execute(context.Background(), ListStockLevelsInput{Actor: admin, Limit: 500})
	assertError(t, err, domain.ErrInvalidPage)

	_, err = NewListStockLevels(&gatewayfake.StockRepository{ListLevelsErr: errDown}).Execute(context.Background(), ListStockLevelsInput{Actor: admin})
	assertError(t, err, domain.ErrStorageUnavailable)
}

func TestListStockMovements(t *testing.T) {
	page := domain.Page[domain.StockMovement]{Items: []domain.StockMovement{{ID: "movement-1"}}, Total: 1, Page: 1, Limit: 25}
	repository := &gatewayfake.StockRepository{ListMovementsResult: page}

	got, err := NewListStockMovements(repository).Execute(context.Background(), ListStockMovementsInput{
		Actor: cashier, ItemID: " yerba ", Type: domain.MovementAdjustment,
	})
	if err != nil || got.Total != 1 {
		t.Fatalf("Execute() = %+v, %v", got, err)
	}
	want := gateway.StockMovementFilter{ItemID: "yerba", Type: domain.MovementAdjustment, Page: domain.PageRequest{Page: 1, Limit: 25}}
	if repository.ListMovementsFilters[0] != want {
		t.Fatalf("filter = %+v, want %+v", repository.ListMovementsFilters[0], want)
	}
}

func TestListStockMovementsErrors(t *testing.T) {
	_, err := NewListStockMovements(&gatewayfake.StockRepository{}).Execute(context.Background(), ListStockMovementsInput{})
	assertError(t, err, domain.ErrForbidden)

	_, err = NewListStockMovements(&gatewayfake.StockRepository{}).Execute(context.Background(), ListStockMovementsInput{Actor: admin, Type: "transfer"})
	assertError(t, err, domain.ErrStockMovementTypeInvalid)

	_, err = NewListStockMovements(&gatewayfake.StockRepository{}).Execute(context.Background(), ListStockMovementsInput{Actor: admin, Page: -1})
	assertError(t, err, domain.ErrInvalidPage)

	_, err = NewListStockMovements(&gatewayfake.StockRepository{ListMovementsErr: errDown}).Execute(context.Background(), ListStockMovementsInput{Actor: admin})
	assertError(t, err, domain.ErrStorageUnavailable)
}
