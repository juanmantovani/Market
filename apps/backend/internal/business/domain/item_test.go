package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var itemTime = time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)

func validItemFields() ItemFields {
	return ItemFields{
		Code:        "  YER-1KG ",
		Barcode:     ptr(" 7790387000123 "),
		Name:        " Yerba mate 1 kg ",
		Description: ptr("  Con palo  "),
		CategoryID:  " category-1 ",
		Unit:        UnitPiece,
		SalePrice:   350_000,
		Cost:        ptr(Money(210_000)),
		TaxRate:     2100,
		MinStock:    WholeQuantity(5),
	}
}

func TestNewItem(t *testing.T) {
	fields := validItemFields()

	item, err := NewItem(fields, itemTime)
	if err != nil {
		t.Fatalf("NewItem() error = %v", err)
	}

	if item.Code != "YER-1KG" || item.Name != "Yerba mate 1 kg" || item.CategoryID != "category-1" {
		t.Errorf("text fields not trimmed: %+v", item.ItemFields)
	}
	if item.Barcode == nil || *item.Barcode != "7790387000123" {
		t.Errorf("Barcode = %v", item.Barcode)
	}
	if item.Description == nil || *item.Description != "Con palo" {
		t.Errorf("Description = %v", item.Description)
	}
	if item.Cost == nil || *item.Cost != 210_000 || item.Cost == fields.Cost {
		t.Errorf("Cost = %v, want a copy of 210000", item.Cost)
	}
	if !item.CreatedAt.Equal(itemTime) || !item.UpdatedAt.Equal(itemTime) || item.ID != "" {
		t.Errorf("item = %+v", item)
	}
}

func TestNewItemAcceptsOptionalFieldsAbsent(t *testing.T) {
	fields := validItemFields()
	fields.Barcode = ptr("   ")
	fields.Description = nil
	fields.Cost = nil
	fields.TaxRate = 0
	fields.MinStock = 0

	item, err := NewItem(fields, itemTime)
	if err != nil {
		t.Fatalf("NewItem() error = %v", err)
	}
	if item.Barcode != nil || item.Description != nil || item.Cost != nil {
		t.Fatalf("optional fields = %v %v %v, want nil", item.Barcode, item.Description, item.Cost)
	}

	fields.Barcode = nil
	if item, err := NewItem(fields, itemTime); err != nil || item.Barcode != nil {
		t.Fatalf("NewItem(no barcode) = %v, %v", item.Barcode, err)
	}
}

func TestNewItemAcceptsFractionalMinStockForWeightedItems(t *testing.T) {
	fields := validItemFields()
	fields.Unit = UnitKilogram
	fields.MinStock = 2500

	if _, err := NewItem(fields, itemTime); err != nil {
		t.Fatalf("NewItem() error = %v", err)
	}
}

func TestNewItemValidates(t *testing.T) {
	tests := []struct {
		name   string
		change func(*ItemFields)
		want   *Error
	}{
		{"missing code", func(f *ItemFields) { f.Code = " " }, ErrItemCodeRequired},
		{"long code", func(f *ItemFields) { f.Code = strings.Repeat("A", MaxItemCodeLength+1) }, ErrItemCodeTooLong},
		{"barcode with spaces", func(f *ItemFields) { f.Barcode = ptr("779 0387") }, ErrItemBarcodeInvalid},
		{"long barcode", func(f *ItemFields) { f.Barcode = ptr(strings.Repeat("7", MaxItemBarcodeLength+1)) }, ErrItemBarcodeInvalid},
		{"missing name", func(f *ItemFields) { f.Name = "" }, ErrItemNameRequired},
		{"long name", func(f *ItemFields) { f.Name = strings.Repeat("n", MaxItemNameLength+1) }, ErrItemNameTooLong},
		{"long description", func(f *ItemFields) { f.Description = ptr(strings.Repeat("d", MaxItemDescriptionLength+1)) }, ErrItemDescriptionTooLong},
		{"missing category", func(f *ItemFields) { f.CategoryID = "  " }, ErrItemCategoryRequired},
		{"unknown unit", func(f *ItemFields) { f.Unit = "box" }, ErrItemUnitInvalid},
		{"zero price", func(f *ItemFields) { f.SalePrice = 0 }, ErrItemPriceInvalid},
		{"huge price", func(f *ItemFields) { f.SalePrice = MaxItemAmount + 1 }, ErrItemPriceInvalid},
		{"negative cost", func(f *ItemFields) { f.Cost = ptr(Money(-1)) }, ErrItemCostInvalid},
		{"huge cost", func(f *ItemFields) { f.Cost = ptr(MaxItemAmount + 1) }, ErrItemCostInvalid},
		{"unknown tax rate", func(f *ItemFields) { f.TaxRate = 1900 }, ErrItemTaxRateInvalid},
		{"negative min stock", func(f *ItemFields) { f.MinStock = -1 }, ErrItemMinStockInvalid},
		{"fractional min stock by the piece", func(f *ItemFields) { f.MinStock = 1500 }, ErrItemMinStockInvalid},
		{"huge min stock", func(f *ItemFields) { f.MinStock = MaxItemMinStock + QuantityScale }, ErrItemMinStockInvalid},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fields := validItemFields()
			test.change(&fields)

			if _, err := NewItem(fields, itemTime); !errors.Is(err, test.want) {
				t.Fatalf("NewItem() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestItemReplace(t *testing.T) {
	original, _ := NewItem(validItemFields(), itemTime)
	original.ID = "item-1"
	later := itemTime.Add(time.Hour)

	fields := validItemFields()
	fields.Name = "Yerba mate suave 1 kg"
	fields.Cost = nil
	replaced, err := original.Replace(fields, later)
	if err != nil {
		t.Fatalf("Replace() error = %v", err)
	}

	if replaced.ID != "item-1" || replaced.Name != "Yerba mate suave 1 kg" || replaced.Cost != nil {
		t.Fatalf("replaced = %+v", replaced)
	}
	if !replaced.CreatedAt.Equal(itemTime) || !replaced.UpdatedAt.Equal(later) {
		t.Fatalf("timestamps = %v / %v", replaced.CreatedAt, replaced.UpdatedAt)
	}
	if original.Cost == nil || original.Name != "Yerba mate 1 kg" {
		t.Fatal("Replace mutated the original item")
	}

	fields.SalePrice = 0
	if _, err := original.Replace(fields, later); !errors.Is(err, ErrItemPriceInvalid) {
		t.Fatalf("Replace(invalid) error = %v, want ErrItemPriceInvalid", err)
	}
}

func TestItemCodeKey(t *testing.T) {
	if got := (Item{ItemFields: ItemFields{Code: " yer-1kg "}}).CodeKey(); got != "yer-1kg" {
		t.Fatalf("CodeKey() = %q", got)
	}
}

func TestNewItemCleansTheSupplierIDs(t *testing.T) {
	fields := validItemFields()
	fields.SupplierIDs = []string{" s1 ", "s2", "", "s1"}

	item, err := NewItem(fields, itemTime)
	if err != nil {
		t.Fatalf("NewItem() error = %v", err)
	}
	if len(item.SupplierIDs) != 2 || item.SupplierIDs[0] != "s1" || item.SupplierIDs[1] != "s2" {
		t.Fatalf("SupplierIDs = %q, want [s1 s2]", item.SupplierIDs)
	}

	fields.SupplierIDs = nil
	item, _ = NewItem(fields, itemTime)
	if item.SupplierIDs == nil || len(item.SupplierIDs) != 0 {
		t.Fatalf("SupplierIDs = %#v, want an empty list", item.SupplierIDs)
	}
}
