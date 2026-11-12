package domain

import (
	"strings"
	"time"
)

const (
	// MaxStockQuantity caps a single movement at a million units, far above
	// any real delivery or count.
	MaxStockQuantity Quantity = 1_000_000 * QuantityScale
	// MaxStockBalance caps the stock of an item at a thousand million units,
	// so adding movements never overflows.
	MaxStockBalance Quantity = 1_000_000_000 * QuantityScale

	MaxStockEntryLines      = 100
	MaxStockReasonLength    = 200
	MaxStockReferenceLength = 60
)

// MovementType says why the stock of an item changed.
type MovementType string

const (
	// MovementInitial is the opening balance, only allowed as the first
	// movement of an item.
	MovementInitial    MovementType = "initial"
	MovementEntry      MovementType = "entry"
	MovementAdjustment MovementType = "adjustment"
	MovementSale       MovementType = "sale"
	MovementSaleVoid   MovementType = "sale_void"
)

func (movementType MovementType) IsValid() bool {
	switch movementType {
	case MovementInitial, MovementEntry, MovementAdjustment, MovementSale, MovementSaleVoid:
		return true
	default:
		return false
	}
}

var (
	ErrStockItemRequired              = &Error{Kind: KindInvalid, Code: "stock_item_required", Message: "Elegir el artículo"}
	ErrStockItemNotFound              = &Error{Kind: KindInvalid, Code: "stock_item_not_found", Message: "El artículo elegido no existe"}
	ErrStockSupplierNotFound          = &Error{Kind: KindInvalid, Code: "stock_supplier_not_found", Message: "El proveedor elegido no existe"}
	ErrStockQuantityInvalid           = &Error{Kind: KindInvalid, Code: "stock_quantity_invalid", Message: "La cantidad tiene que ser mayor a cero y de hasta un millón de unidades"}
	ErrStockAdjustmentQuantityInvalid = &Error{Kind: KindInvalid, Code: "stock_adjustment_quantity_invalid", Message: "La cantidad no puede ser cero ni pasar del millón de unidades"}
	ErrStockQuantityNotWhole          = &Error{Kind: KindInvalid, Code: "stock_quantity_not_whole", Message: "El artículo se vende por unidad: la cantidad tiene que ser entera"}
	ErrStockBalanceOutOfRange         = &Error{Kind: KindInvalid, Code: "stock_balance_out_of_range", Message: "El stock resultante excede el máximo permitido"}
	ErrStockUnitCostInvalid           = &Error{Kind: KindInvalid, Code: "stock_unit_cost_invalid", Message: "El costo unitario no puede ser negativo"}
	ErrStockEntryLinesRequired        = &Error{Kind: KindInvalid, Code: "stock_entry_lines_required", Message: "Agregar al menos un artículo"}
	ErrStockEntryTooManyLines         = &Error{Kind: KindInvalid, Code: "stock_entry_too_many_lines", Message: "El ingreso admite hasta 100 artículos"}
	ErrStockEntryItemRepeated         = &Error{Kind: KindInvalid, Code: "stock_entry_item_repeated", Message: "Un artículo aparece en más de una línea: sumar las cantidades en una sola"}
	ErrStockReferenceTooLong          = &Error{Kind: KindInvalid, Code: "stock_reference_too_long", Message: "El comprobante admite hasta 60 caracteres"}
	ErrStockReasonRequired            = &Error{Kind: KindInvalid, Code: "stock_reason_required", Message: "Ingresar el motivo del ajuste"}
	ErrStockReasonTooLong             = &Error{Kind: KindInvalid, Code: "stock_reason_too_long", Message: "El motivo admite hasta 200 caracteres"}
	ErrStockMovementTypeInvalid       = &Error{Kind: KindInvalid, Code: "stock_movement_type_invalid", Message: "El tipo de movimiento no es válido"}
	ErrStockAlreadyInitialized        = &Error{Kind: KindConflict, Code: "stock_already_initialized", Message: "El artículo ya tiene movimientos de stock"}
	ErrInsufficientStock              = &Error{Kind: KindConflict, Code: "insufficient_stock", Message: "No hay stock suficiente"}
	ErrItemInUse                      = &Error{Kind: KindConflict, Code: "item_in_use", Message: "El artículo tiene movimientos de stock y no se puede eliminar"}
	// ErrSupplierHasStockEntries shares the code of ErrSupplierInUse, so
	// errors.Is matches both; only the message says why.
	ErrSupplierHasStockEntries = &Error{Kind: KindConflict, Code: "supplier_in_use", Message: "El proveedor tiene ingresos de mercadería registrados y no se puede eliminar"}
)

// InsufficientStock is ErrInsufficientStock naming the item that ran short.
func InsufficientStock(itemName string) *Error {
	err := *ErrInsufficientStock
	err.Message = "No hay stock suficiente de " + itemName

	return &err
}

// IsLowStock reports whether quantity reached the minimum of the item. A
// zero minimum turns the alert off.
func IsLowStock(quantity, minStock Quantity) bool {
	return minStock > 0 && quantity <= minStock
}

// StockLevel is how much there is of an item. It only changes through a
// movement, and an item without movements has a zero level.
type StockLevel struct {
	ItemID   string
	Quantity Quantity
	// UpdatedAt is when the last movement happened; zero while there is none.
	UpdatedAt time.Time
}

func (level StockLevel) HasMovements() bool {
	return !level.UpdatedAt.IsZero()
}

// ItemStock is an item together with its stock level.
type ItemStock struct {
	Item  Item
	Level StockLevel
}

func (itemStock ItemStock) IsLow() bool {
	return IsLowStock(itemStock.Level.Quantity, itemStock.Item.MinStock)
}

// StockMovement is one change in the stock of an item. Movements are never
// edited or deleted: a mistake is undone with another movement.
type StockMovement struct {
	ID     string
	ItemID string
	Type   MovementType
	// Quantity is signed: what came in is positive, what went out negative.
	Quantity     Quantity
	BalanceAfter Quantity
	Reason       *string
	Reference    *string
	SupplierID   *string
	UnitCost     *Money
	ActorID      string
	CreatedAt    time.Time
}

// StockEntryLine is one item received in a stock entry.
type StockEntryLine struct {
	ItemID   string
	Quantity Quantity
	UnitCost *Money
}

// StockEntry is merchandise received, with one line per item. The supplier
// and the reference (a delivery note or an invoice) are optional.
type StockEntry struct {
	SupplierID *string
	Reference  *string
	Lines      []StockEntryLine
}

// NewStockEntry validates what an entry must satisfy before looking at any
// stored data. The result never aliases the caller's pointers or slice.
func NewStockEntry(entry StockEntry) (StockEntry, error) {
	reference, err := optionalText(entry.Reference, MaxStockReferenceLength, ErrStockReferenceTooLong)
	if err != nil {
		return StockEntry{}, err
	}
	if len(entry.Lines) == 0 {
		return StockEntry{}, ErrStockEntryLinesRequired
	}
	if len(entry.Lines) > MaxStockEntryLines {
		return StockEntry{}, ErrStockEntryTooManyLines
	}

	valid := StockEntry{SupplierID: optionalID(entry.SupplierID), Reference: reference, Lines: make([]StockEntryLine, 0, len(entry.Lines))}
	seen := make(map[string]bool, len(entry.Lines))
	for _, line := range entry.Lines {
		itemID := strings.TrimSpace(line.ItemID)
		if itemID == "" {
			return StockEntry{}, ErrStockItemRequired
		}
		if seen[itemID] {
			return StockEntry{}, ErrStockEntryItemRepeated
		}
		seen[itemID] = true
		if line.Quantity <= 0 || line.Quantity > MaxStockQuantity {
			return StockEntry{}, ErrStockQuantityInvalid
		}
		validLine := StockEntryLine{ItemID: itemID, Quantity: line.Quantity}
		if line.UnitCost != nil {
			if *line.UnitCost < 0 || *line.UnitCost > MaxItemAmount {
				return StockEntry{}, ErrStockUnitCostInvalid
			}
			cost := *line.UnitCost
			validLine.UnitCost = &cost
		}
		valid.Lines = append(valid.Lines, validLine)
	}

	return valid, nil
}

// UpdateCost returns item with the unit cost paid in the line as its cost,
// and whether that changed it: an entry with a cost updates the item.
func (line StockEntryLine) UpdateCost(item Item, now time.Time) (Item, bool, error) {
	if line.UnitCost == nil || (item.Cost != nil && *item.Cost == *line.UnitCost) {
		return item, false, nil
	}

	fields := item.ItemFields
	cost := *line.UnitCost
	fields.Cost = &cost
	updated, err := item.Replace(fields, now)
	if err != nil {
		return Item{}, false, err
	}

	return updated, true, nil
}

// StockAdjustment is a manual correction of the stock of an item, positive
// or negative, always with a reason.
type StockAdjustment struct {
	ItemID   string
	Quantity Quantity
	Reason   string
}

// NewStockAdjustment validates what an adjustment must satisfy before
// looking at any stored data.
func NewStockAdjustment(adjustment StockAdjustment) (StockAdjustment, error) {
	valid := adjustment
	if valid.ItemID = strings.TrimSpace(adjustment.ItemID); valid.ItemID == "" {
		return StockAdjustment{}, ErrStockItemRequired
	}
	if adjustment.Quantity == 0 || adjustment.Quantity > MaxStockQuantity || adjustment.Quantity < -MaxStockQuantity {
		return StockAdjustment{}, ErrStockAdjustmentQuantityInvalid
	}
	reason, err := requiredText(adjustment.Reason, MaxStockReasonLength, ErrStockReasonRequired, ErrStockReasonTooLong)
	if err != nil {
		return StockAdjustment{}, err
	}
	valid.Reason = reason

	return valid, nil
}

// Initial is the movement that opens the stock of item. It is only allowed
// while the item has no movements.
func (level StockLevel) Initial(item Item, quantity Quantity, actor Actor, now time.Time) (StockMovement, error) {
	if level.HasMovements() {
		return StockMovement{}, ErrStockAlreadyInitialized
	}
	if quantity <= 0 || quantity > MaxStockQuantity {
		return StockMovement{}, ErrStockQuantityInvalid
	}

	return level.move(item.ID, item.Unit, MovementInitial, quantity, actor, now)
}

// Receive is the movement that adds one line of a validated entry to the
// stock of its item.
func (level StockLevel) Receive(item Item, entry StockEntry, line StockEntryLine, actor Actor, now time.Time) (StockMovement, error) {
	movement, err := level.move(item.ID, item.Unit, MovementEntry, line.Quantity, actor, now)
	if err != nil {
		return StockMovement{}, err
	}
	movement.SupplierID = entry.SupplierID
	movement.Reference = entry.Reference
	movement.UnitCost = line.UnitCost

	return movement, nil
}

// Adjust is the movement that corrects the stock of item with a validated
// adjustment. Taking stock out can't leave it below zero, since nobody counts
// a negative quantity on the shelf; adding always works, even to a stock
// that sales left negative.
func (level StockLevel) Adjust(item Item, adjustment StockAdjustment, actor Actor, now time.Time) (StockMovement, error) {
	movement, err := level.move(item.ID, item.Unit, MovementAdjustment, adjustment.Quantity, actor, now)
	if err != nil {
		return StockMovement{}, err
	}
	if adjustment.Quantity < 0 && movement.BalanceAfter < 0 {
		return StockMovement{}, InsufficientStock(item.Name)
	}
	reason := adjustment.Reason
	movement.Reason = &reason

	return movement, nil
}

// StockPolicy is how sales treat the stock.
type StockPolicy struct {
	// AllowNegative lets a sale take more than there is, for stores that
	// prefer not to stop the till: the stock goes below zero until an entry
	// or an adjustment fixes it.
	AllowNegative bool
}

// Sell is the movement that takes one line of sale out of the stock of its
// item. Unless policy allows a negative stock, a line the stock doesn't
// cover fails with insufficient_stock naming the item.
func (level StockLevel) Sell(sale Sale, line SaleLine, policy StockPolicy) (StockMovement, error) {
	movement, err := level.move(line.ItemID, line.Unit, MovementSale, -line.Quantity, Actor{ID: sale.ActorID}, sale.CreatedAt)
	if err != nil {
		return StockMovement{}, err
	}
	if !policy.AllowNegative && movement.BalanceAfter < 0 {
		return StockMovement{}, InsufficientStock(line.Name)
	}
	reference := sale.Reference()
	movement.Reference = &reference

	return movement, nil
}

// move is the movement that changes the level of an item by quantity.
func (level StockLevel) move(itemID string, unit Unit, movementType MovementType, quantity Quantity, actor Actor, now time.Time) (StockMovement, error) {
	if !unit.AllowsFraction() && !quantity.IsWhole() {
		return StockMovement{}, ErrStockQuantityNotWhole
	}
	balance := level.Quantity + quantity
	if balance > MaxStockBalance || balance < -MaxStockBalance {
		return StockMovement{}, ErrStockBalanceOutOfRange
	}

	return StockMovement{
		ItemID:       itemID,
		Type:         movementType,
		Quantity:     quantity,
		BalanceAfter: balance,
		ActorID:      actor.ID,
		CreatedAt:    now,
	}, nil
}

// optionalID trims an optional ID, treating a blank one as absent. The
// result never aliases the caller's pointer.
func optionalID(id *string) *string {
	if id == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*id)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
