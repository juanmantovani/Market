package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var (
	stockTime  = time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)
	stockActor = Actor{ID: "u-manager", Role: RoleManager}
)

func stockItem(unit Unit) Item {
	return Item{ID: "item-1", ItemFields: ItemFields{
		Code: "YER-1", Name: "Yerba", CategoryID: "category-1", Unit: unit, SalePrice: 100_000, TaxRate: 2100,
	}}
}

func money(amount Money) *Money {
	return &amount
}

func textPtr(value string) *string {
	return &value
}

func assertDomainError(t *testing.T, err error, want *Error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %s", err, want.Code)
	}
}

func TestMovementTypeIsValid(t *testing.T) {
	for _, movementType := range []MovementType{MovementInitial, MovementEntry, MovementAdjustment, MovementSale, MovementSaleVoid} {
		if !movementType.IsValid() {
			t.Errorf("%q.IsValid() = false", movementType)
		}
	}
	if MovementType("transfer").IsValid() {
		t.Error(`"transfer".IsValid() = true`)
	}
}

func TestIsLowStock(t *testing.T) {
	tests := []struct {
		quantity, minStock Quantity
		want               bool
	}{
		{quantity: 3000, minStock: 5000, want: true},
		{quantity: 5000, minStock: 5000, want: true},
		{quantity: 5001, minStock: 5000, want: false},
		{quantity: -1000, minStock: 1000, want: true},
		{quantity: 0, minStock: 0, want: false},
	}
	for _, test := range tests {
		if got := IsLowStock(test.quantity, test.minStock); got != test.want {
			t.Errorf("IsLowStock(%d, %d) = %v, want %v", test.quantity, test.minStock, got, test.want)
		}
	}

	item := stockItem(UnitPiece)
	item.MinStock = WholeQuantity(6)
	if !(ItemStock{Item: item, Level: StockLevel{Quantity: WholeQuantity(4)}}).IsLow() {
		t.Error("4 of a minimum of 6 is not low")
	}
	if (ItemStock{Item: item, Level: StockLevel{Quantity: WholeQuantity(7)}}).IsLow() {
		t.Error("7 of a minimum of 6 is low")
	}
}

func TestInsufficientStockNamesTheItem(t *testing.T) {
	err := InsufficientStock("Yerba 1 kg")

	if err.Message != "No hay stock suficiente de Yerba 1 kg" || !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("InsufficientStock() = %+v", err)
	}
	if ErrInsufficientStock.Message != "No hay stock suficiente" {
		t.Fatal("InsufficientStock changed the sentinel")
	}
}

func TestSupplierHasStockEntriesIsASupplierInUse(t *testing.T) {
	if !errors.Is(ErrSupplierHasStockEntries, ErrSupplierInUse) {
		t.Fatal("ErrSupplierHasStockEntries does not match ErrSupplierInUse")
	}
}

func TestNewStockEntry(t *testing.T) {
	supplierID, reference := " supplier-1 ", " Remito 0001-00001234 "
	cost := Money(250_000)
	input := StockEntry{
		SupplierID: &supplierID,
		Reference:  &reference,
		Lines: []StockEntryLine{
			{ItemID: " item-1 ", Quantity: WholeQuantity(12), UnitCost: &cost},
			{ItemID: "item-2", Quantity: 1500},
		},
	}

	entry, err := NewStockEntry(input)
	if err != nil {
		t.Fatalf("NewStockEntry() error = %v", err)
	}

	if *entry.SupplierID != "supplier-1" || *entry.Reference != "Remito 0001-00001234" || len(entry.Lines) != 2 {
		t.Fatalf("entry = %+v", entry)
	}
	if entry.Lines[0].ItemID != "item-1" || entry.Lines[0].Quantity != 12_000 || *entry.Lines[0].UnitCost != 250_000 || entry.Lines[1].UnitCost != nil {
		t.Fatalf("lines = %+v", entry.Lines)
	}
	cost, supplierID = 1, "changed"
	if *entry.Lines[0].UnitCost != 250_000 || *entry.SupplierID != "supplier-1" {
		t.Fatal("the entry aliases the caller's pointers")
	}

	blank := " "
	withoutSupplier, err := NewStockEntry(StockEntry{SupplierID: &blank, Reference: &blank, Lines: input.Lines})
	if err != nil || withoutSupplier.SupplierID != nil || withoutSupplier.Reference != nil {
		t.Fatalf("NewStockEntry(blank supplier and reference) = %+v, %v", withoutSupplier, err)
	}
}

func TestNewStockEntryRejects(t *testing.T) {
	line := func(itemID string, quantity Quantity) StockEntryLine {
		return StockEntryLine{ItemID: itemID, Quantity: quantity}
	}
	tooMany := make([]StockEntryLine, MaxStockEntryLines+1)
	for i := range tooMany {
		tooMany[i] = line("item-"+strings.Repeat("x", i), 1000)
	}
	longReference := strings.Repeat("r", MaxStockReferenceLength+1)

	tests := []struct {
		name  string
		entry StockEntry
		want  *Error
	}{
		{"no lines", StockEntry{}, ErrStockEntryLinesRequired},
		{"too many lines", StockEntry{Lines: tooMany}, ErrStockEntryTooManyLines},
		{"long reference", StockEntry{Reference: &longReference, Lines: []StockEntryLine{line("item-1", 1000)}}, ErrStockReferenceTooLong},
		{"blank item", StockEntry{Lines: []StockEntryLine{line(" ", 1000)}}, ErrStockItemRequired},
		{"repeated item", StockEntry{Lines: []StockEntryLine{line("item-1", 1000), line(" item-1", 2000)}}, ErrStockEntryItemRepeated},
		{"zero quantity", StockEntry{Lines: []StockEntryLine{line("item-1", 0)}}, ErrStockQuantityInvalid},
		{"negative quantity", StockEntry{Lines: []StockEntryLine{line("item-1", -1000)}}, ErrStockQuantityInvalid},
		{"huge quantity", StockEntry{Lines: []StockEntryLine{line("item-1", MaxStockQuantity+1)}}, ErrStockQuantityInvalid},
		{"negative cost", StockEntry{Lines: []StockEntryLine{{ItemID: "item-1", Quantity: 1000, UnitCost: money(-1)}}}, ErrStockUnitCostInvalid},
		{"huge cost", StockEntry{Lines: []StockEntryLine{{ItemID: "item-1", Quantity: 1000, UnitCost: money(MaxItemAmount + 1)}}}, ErrStockUnitCostInvalid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewStockEntry(test.entry)
			assertDomainError(t, err, test.want)
		})
	}

	if _, err := NewStockEntry(StockEntry{Lines: []StockEntryLine{line("item-1", MaxStockQuantity)}}); err != nil {
		t.Fatalf("NewStockEntry(max quantity) error = %v", err)
	}
	if _, err := NewStockEntry(StockEntry{Lines: tooMany[:MaxStockEntryLines]}); err != nil {
		t.Fatalf("NewStockEntry(max lines) error = %v", err)
	}
}

func TestStockEntryLineUpdateCost(t *testing.T) {
	item := stockItem(UnitPiece)
	item.Cost = money(200_000)
	item.SupplierIDs = []string{"supplier-1"}

	unchanged, changed, err := StockEntryLine{UnitCost: nil}.UpdateCost(item, stockTime)
	if err != nil || changed || *unchanged.Cost != 200_000 {
		t.Fatalf("UpdateCost(no cost) = %+v, %v, %v", unchanged, changed, err)
	}
	unchanged, changed, err = StockEntryLine{UnitCost: money(200_000)}.UpdateCost(item, stockTime)
	if err != nil || changed || !unchanged.UpdatedAt.IsZero() {
		t.Fatalf("UpdateCost(same cost) = %+v, %v, %v", unchanged, changed, err)
	}

	updated, changed, err := StockEntryLine{UnitCost: money(230_000)}.UpdateCost(item, stockTime)
	if err != nil || !changed || *updated.Cost != 230_000 || !updated.UpdatedAt.Equal(stockTime) || updated.SupplierIDs[0] != "supplier-1" {
		t.Fatalf("UpdateCost(new cost) = %+v, %v, %v", updated, changed, err)
	}
	if *item.Cost != 200_000 {
		t.Fatal("UpdateCost changed the caller's item")
	}

	withoutCost := stockItem(UnitPiece)
	updated, changed, err = StockEntryLine{UnitCost: money(0)}.UpdateCost(withoutCost, stockTime)
	if err != nil || !changed || *updated.Cost != 0 {
		t.Fatalf("UpdateCost(first cost) = %+v, %v, %v", updated, changed, err)
	}

	broken := stockItem("box")
	if _, _, err := (StockEntryLine{UnitCost: money(1)}).UpdateCost(broken, stockTime); !errors.Is(err, ErrItemUnitInvalid) {
		t.Fatalf("UpdateCost(invalid item) error = %v, want ErrItemUnitInvalid", err)
	}
}

func TestNewStockAdjustment(t *testing.T) {
	adjustment, err := NewStockAdjustment(StockAdjustment{ItemID: " item-1 ", Quantity: -2000, Reason: "  Rotura  "})
	if err != nil {
		t.Fatalf("NewStockAdjustment() error = %v", err)
	}
	if adjustment.ItemID != "item-1" || adjustment.Quantity != -2000 || adjustment.Reason != "Rotura" {
		t.Fatalf("adjustment = %+v", adjustment)
	}

	tests := []struct {
		name       string
		adjustment StockAdjustment
		want       *Error
	}{
		{"blank item", StockAdjustment{ItemID: " ", Quantity: 1000, Reason: "Conteo"}, ErrStockItemRequired},
		{"zero", StockAdjustment{ItemID: "item-1", Reason: "Conteo"}, ErrStockAdjustmentQuantityInvalid},
		{"too much in", StockAdjustment{ItemID: "item-1", Quantity: MaxStockQuantity + 1, Reason: "Conteo"}, ErrStockAdjustmentQuantityInvalid},
		{"too much out", StockAdjustment{ItemID: "item-1", Quantity: -MaxStockQuantity - 1, Reason: "Conteo"}, ErrStockAdjustmentQuantityInvalid},
		{"no reason", StockAdjustment{ItemID: "item-1", Quantity: 1000, Reason: "  "}, ErrStockReasonRequired},
		{"long reason", StockAdjustment{ItemID: "item-1", Quantity: 1000, Reason: strings.Repeat("m", MaxStockReasonLength+1)}, ErrStockReasonTooLong},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewStockAdjustment(test.adjustment)
			assertDomainError(t, err, test.want)
		})
	}
}

func TestStockLevelInitial(t *testing.T) {
	item := stockItem(UnitKilogram)
	empty := StockLevel{ItemID: item.ID}

	movement, err := empty.Initial(item, 2500, stockActor, stockTime)
	if err != nil {
		t.Fatalf("Initial() error = %v", err)
	}
	want := StockMovement{ItemID: "item-1", Type: MovementInitial, Quantity: 2500, BalanceAfter: 2500, ActorID: "u-manager", CreatedAt: stockTime}
	if movement != want {
		t.Fatalf("Initial() = %+v, want %+v", movement, want)
	}

	tests := []struct {
		name     string
		level    StockLevel
		item     Item
		quantity Quantity
		want     *Error
	}{
		{"already has movements", StockLevel{ItemID: item.ID, UpdatedAt: stockTime}, item, 1000, ErrStockAlreadyInitialized},
		{"zero", empty, item, 0, ErrStockQuantityInvalid},
		{"too much", empty, item, MaxStockQuantity + 1, ErrStockQuantityInvalid},
		{"fraction of a unit", empty, stockItem(UnitPiece), 1500, ErrStockQuantityNotWhole},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.level.Initial(test.item, test.quantity, stockActor, stockTime)
			assertDomainError(t, err, test.want)
		})
	}
}

func TestStockLevelReceive(t *testing.T) {
	item := stockItem(UnitPiece)
	entry, err := NewStockEntry(StockEntry{
		SupplierID: textPtr("supplier-1"),
		Reference:  textPtr("Remito 12"),
		Lines:      []StockEntryLine{{ItemID: item.ID, Quantity: WholeQuantity(12), UnitCost: money(250_000)}},
	})
	if err != nil {
		t.Fatalf("NewStockEntry() error = %v", err)
	}
	level := StockLevel{ItemID: item.ID, Quantity: WholeQuantity(3), UpdatedAt: stockTime}

	movement, err := level.Receive(item, entry, entry.Lines[0], stockActor, stockTime)
	if err != nil {
		t.Fatalf("Receive() error = %v", err)
	}
	if movement.Type != MovementEntry || movement.Quantity != 12_000 || movement.BalanceAfter != 15_000 ||
		*movement.SupplierID != "supplier-1" || *movement.Reference != "Remito 12" || *movement.UnitCost != 250_000 ||
		movement.Reason != nil || movement.ActorID != "u-manager" || !movement.CreatedAt.Equal(stockTime) {
		t.Fatalf("Receive() = %+v", movement)
	}

	fraction := StockEntryLine{ItemID: item.ID, Quantity: 500}
	if _, err := level.Receive(item, entry, fraction, stockActor, stockTime); !errors.Is(err, ErrStockQuantityNotWhole) {
		t.Fatalf("Receive(half a unit) error = %v, want ErrStockQuantityNotWhole", err)
	}

	full := StockLevel{ItemID: item.ID, Quantity: MaxStockBalance}
	if _, err := full.Receive(item, entry, entry.Lines[0], stockActor, stockTime); !errors.Is(err, ErrStockBalanceOutOfRange) {
		t.Fatalf("Receive(over the maximum balance) error = %v, want ErrStockBalanceOutOfRange", err)
	}
}

func TestStockLevelAdjust(t *testing.T) {
	item := stockItem(UnitKilogram)
	item.Name = "Queso cremoso"
	adjust := func(quantity Quantity) StockAdjustment {
		adjustment, err := NewStockAdjustment(StockAdjustment{ItemID: item.ID, Quantity: quantity, Reason: "Conteo"})
		if err != nil {
			t.Fatalf("NewStockAdjustment(%d) error = %v", quantity, err)
		}
		return adjustment
	}
	level := StockLevel{ItemID: item.ID, Quantity: 1800, UpdatedAt: stockTime}

	out, err := level.Adjust(item, adjust(-1800), stockActor, stockTime)
	if err != nil || out.Type != MovementAdjustment || out.Quantity != -1800 || out.BalanceAfter != 0 || *out.Reason != "Conteo" {
		t.Fatalf("Adjust(-1.8 kg) = %+v, %v", out, err)
	}
	in, err := level.Adjust(item, adjust(250), stockActor, stockTime)
	if err != nil || in.BalanceAfter != 2050 || in.SupplierID != nil || in.UnitCost != nil {
		t.Fatalf("Adjust(+0.25 kg) = %+v, %v", in, err)
	}

	_, err = level.Adjust(item, adjust(-1801), stockActor, stockTime)
	assertDomainError(t, err, ErrInsufficientStock)
	if !strings.Contains(err.Error(), "Queso cremoso") {
		t.Fatalf("Adjust(below zero) message = %q, want the item name", err.Error())
	}

	negative := StockLevel{ItemID: item.ID, Quantity: -3000, UpdatedAt: stockTime}
	if up, err := negative.Adjust(item, adjust(1000), stockActor, stockTime); err != nil || up.BalanceAfter != -2000 {
		t.Fatalf("Adjust(+1 kg on a negative stock) = %+v, %v", up, err)
	}

	pieces := stockItem(UnitPiece)
	if _, err := level.Adjust(pieces, adjust(-500), stockActor, stockTime); !errors.Is(err, ErrStockQuantityNotWhole) {
		t.Fatalf("Adjust(half a unit) error = %v, want ErrStockQuantityNotWhole", err)
	}

	bottom := StockLevel{ItemID: item.ID, Quantity: -MaxStockBalance}
	if _, err := bottom.Adjust(item, adjust(-1000), stockActor, stockTime); !errors.Is(err, ErrStockBalanceOutOfRange) {
		t.Fatalf("Adjust(under the minimum balance) error = %v, want ErrStockBalanceOutOfRange", err)
	}
}
