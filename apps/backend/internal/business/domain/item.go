package domain

import (
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	MaxItemCodeLength        = 30
	MaxItemBarcodeLength     = 32
	MaxItemNameLength        = 100
	MaxItemDescriptionLength = 500

	// MaxItemAmount caps prices and costs at a thousand million pesos, far
	// above any real price, so totals of many lines never overflow int64.
	MaxItemAmount Money = 1_000_000_000_00
	// MaxItemMinStock caps the low stock threshold at a million units.
	MaxItemMinStock Quantity = 1_000_000 * QuantityScale
)

// TaxRates are the IVA rates an item can carry. Prices include IVA; the
// rate is informative until electronic invoicing arrives.
var TaxRates = []Percent{0, 1050, 2100, 2700}

var (
	ErrItemNotFound           = &Error{Kind: KindNotFound, Code: "item_not_found", Message: "El artículo no existe"}
	ErrItemCodeRequired       = &Error{Kind: KindInvalid, Code: "item_code_required", Message: "El código es obligatorio"}
	ErrItemCodeTooLong        = &Error{Kind: KindInvalid, Code: "item_code_too_long", Message: "El código admite hasta 30 caracteres"}
	ErrItemCodeInUse          = &Error{Kind: KindConflict, Code: "item_code_in_use", Message: "Ya hay un artículo con ese código"}
	ErrItemBarcodeInvalid     = &Error{Kind: KindInvalid, Code: "item_barcode_invalid", Message: "El código de barras no puede tener espacios y admite hasta 32 caracteres"}
	ErrItemBarcodeInUse       = &Error{Kind: KindConflict, Code: "item_barcode_in_use", Message: "Ya hay un artículo con ese código de barras"}
	ErrItemNameRequired       = &Error{Kind: KindInvalid, Code: "item_name_required", Message: "El nombre es obligatorio"}
	ErrItemNameTooLong        = &Error{Kind: KindInvalid, Code: "item_name_too_long", Message: "El nombre admite hasta 100 caracteres"}
	ErrItemDescriptionTooLong = &Error{Kind: KindInvalid, Code: "item_description_too_long", Message: "La descripción admite hasta 500 caracteres"}
	ErrItemCategoryRequired   = &Error{Kind: KindInvalid, Code: "item_category_required", Message: "Elegir una categoría"}
	ErrItemCategoryNotFound   = &Error{Kind: KindInvalid, Code: "item_category_not_found", Message: "La categoría elegida no existe"}
	ErrItemSupplierNotFound   = &Error{Kind: KindInvalid, Code: "item_supplier_not_found", Message: "Uno de los proveedores elegidos no existe"}
	ErrItemUnitInvalid        = &Error{Kind: KindInvalid, Code: "item_unit_invalid", Message: "La unidad no es válida"}
	ErrItemPriceInvalid       = &Error{Kind: KindInvalid, Code: "item_price_invalid", Message: "El precio tiene que ser mayor a cero"}
	ErrItemCostInvalid        = &Error{Kind: KindInvalid, Code: "item_cost_invalid", Message: "El costo no puede ser negativo"}
	ErrItemTaxRateInvalid     = &Error{Kind: KindInvalid, Code: "item_tax_rate_invalid", Message: "La alícuota de IVA no es válida"}
	ErrItemMinStockInvalid    = &Error{Kind: KindInvalid, Code: "item_min_stock_invalid", Message: "El stock mínimo no es válido para la unidad del artículo"}
	ErrCategoryInUse          = &Error{Kind: KindConflict, Code: "category_in_use", Message: "La categoría tiene artículos: cambiar la categoría de esos artículos antes de eliminarla"}
)

// ItemFields are the editable fields of an item. Creating and editing both
// take the whole set, so an optional field is cleared by leaving it empty.
type ItemFields struct {
	Code        string
	Barcode     *string
	Name        string
	Description *string
	CategoryID  string
	// SupplierIDs are who the item can be bought from, in no particular
	// order; an item may have none.
	SupplierIDs []string
	Unit        Unit
	SalePrice   Money
	Cost        *Money
	TaxRate     Percent
	MinStock    Quantity
}

// Item is a product that can be sold. SalePrice includes IVA.
type Item struct {
	ID string
	ItemFields
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewItem validates the fields of an item about to be created. The
// repository assigns its ID.
func NewItem(fields ItemFields, now time.Time) (Item, error) {
	valid, err := validateItemFields(fields)
	if err != nil {
		return Item{}, err
	}

	return Item{ItemFields: valid, CreatedAt: now, UpdatedAt: now}, nil
}

// Replace returns the item with all its editable fields replaced.
func (item Item) Replace(fields ItemFields, now time.Time) (Item, error) {
	valid, err := validateItemFields(fields)
	if err != nil {
		return Item{}, err
	}

	updated := item
	updated.ItemFields = valid
	updated.UpdatedAt = now

	return updated, nil
}

// CodeKey is the form of the code used to detect duplicates.
func (item Item) CodeKey() string {
	return NormalizeText(item.Code)
}

func validateItemFields(fields ItemFields) (ItemFields, error) {
	var err error
	valid := fields

	if valid.Code, err = requiredText(fields.Code, MaxItemCodeLength, ErrItemCodeRequired, ErrItemCodeTooLong); err != nil {
		return ItemFields{}, err
	}
	if valid.Barcode, err = validateBarcode(fields.Barcode); err != nil {
		return ItemFields{}, err
	}
	if valid.Name, err = requiredText(fields.Name, MaxItemNameLength, ErrItemNameRequired, ErrItemNameTooLong); err != nil {
		return ItemFields{}, err
	}
	if valid.Description, err = optionalText(fields.Description, MaxItemDescriptionLength, ErrItemDescriptionTooLong); err != nil {
		return ItemFields{}, err
	}
	if valid.CategoryID = strings.TrimSpace(fields.CategoryID); valid.CategoryID == "" {
		return ItemFields{}, ErrItemCategoryRequired
	}
	valid.SupplierIDs = uniqueIDs(fields.SupplierIDs)
	if !fields.Unit.IsValid() {
		return ItemFields{}, ErrItemUnitInvalid
	}
	if fields.SalePrice <= 0 || fields.SalePrice > MaxItemAmount {
		return ItemFields{}, ErrItemPriceInvalid
	}
	if fields.Cost != nil {
		if *fields.Cost < 0 || *fields.Cost > MaxItemAmount {
			return ItemFields{}, ErrItemCostInvalid
		}
		cost := *fields.Cost
		valid.Cost = &cost
	}
	if !slices.Contains(TaxRates, fields.TaxRate) {
		return ItemFields{}, ErrItemTaxRateInvalid
	}
	if fields.MinStock < 0 || fields.MinStock > MaxItemMinStock || (!fields.Unit.AllowsFraction() && !fields.MinStock.IsWhole()) {
		return ItemFields{}, ErrItemMinStockInvalid
	}

	return valid, nil
}

// uniqueIDs trims the IDs and drops blanks and repeats, keeping the first
// occurrence. The result is never nil and never aliases the caller's slice.
func uniqueIDs(ids []string) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if id = strings.TrimSpace(id); id != "" && !slices.Contains(result, id) {
			result = append(result, id)
		}
	}

	return result
}

// validateBarcode keeps a barcode as the scanner reads it: no spaces inside,
// at most MaxItemBarcodeLength characters. A blank barcode is absent.
func validateBarcode(barcode *string) (*string, error) {
	if barcode == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*barcode)
	if trimmed == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(trimmed) > MaxItemBarcodeLength || strings.ContainsFunc(trimmed, unicode.IsSpace) {
		return nil, ErrItemBarcodeInvalid
	}

	return &trimmed, nil
}
