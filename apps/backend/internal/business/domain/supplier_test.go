package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var supplierTime = time.Date(2026, 11, 11, 12, 0, 0, 0, time.UTC)

func validSupplierFields() SupplierFields {
	return SupplierFields{
		BusinessName: "  Distribuidora del Centro SA ",
		CUIT:         ptr(" 30-70123456-1 "),
		ContactName:  ptr(" Laura Díaz "),
		Phone:        ptr(" 351 555-2000 "),
		Email:        ptr(" pedidos@delcentro.com.ar "),
		Address:      ptr(" Av. Colón 1500, Córdoba "),
		Notes:        ptr(" Entrega martes y viernes "),
	}
}

func TestNewSupplier(t *testing.T) {
	supplier, err := NewSupplier(validSupplierFields(), supplierTime)
	if err != nil {
		t.Fatalf("NewSupplier() error = %v", err)
	}

	if supplier.BusinessName != "Distribuidora del Centro SA" || *supplier.CUIT != "30701234561" {
		t.Errorf("supplier = %+v", supplier.SupplierFields)
	}
	if *supplier.ContactName != "Laura Díaz" || *supplier.Phone != "351 555-2000" || *supplier.Email != "pedidos@delcentro.com.ar" ||
		*supplier.Address != "Av. Colón 1500, Córdoba" || *supplier.Notes != "Entrega martes y viernes" {
		t.Errorf("optional fields not trimmed: %+v", supplier.SupplierFields)
	}
	if !supplier.CreatedAt.Equal(supplierTime) || !supplier.UpdatedAt.Equal(supplierTime) || supplier.ID != "" {
		t.Errorf("supplier = %+v", supplier)
	}
}

func TestNewSupplierWithOnlyABusinessName(t *testing.T) {
	supplier, err := NewSupplier(SupplierFields{BusinessName: "Verdulería Don José", CUIT: ptr(" "), Email: ptr("")}, supplierTime)
	if err != nil {
		t.Fatalf("NewSupplier() error = %v", err)
	}

	if supplier.CUIT != nil || supplier.Email != nil || supplier.ContactName != nil {
		t.Fatalf("supplier = %+v", supplier.SupplierFields)
	}
}

func TestNewSupplierValidates(t *testing.T) {
	tests := []struct {
		name   string
		change func(*SupplierFields)
		want   *Error
	}{
		{"missing business name", func(f *SupplierFields) { f.BusinessName = " " }, ErrSupplierBusinessNameRequired},
		{"long business name", func(f *SupplierFields) { f.BusinessName = strings.Repeat("n", MaxSupplierBusinessNameLength+1) }, ErrSupplierBusinessNameTooLong},
		{"short CUIT", func(f *SupplierFields) { f.CUIT = ptr("30-7012345-1") }, ErrSupplierCUITInvalid},
		{"CUIT with letters", func(f *SupplierFields) { f.CUIT = ptr("30-7012345A-1") }, ErrSupplierCUITInvalid},
		{"long contact name", func(f *SupplierFields) { f.ContactName = ptr(strings.Repeat("n", MaxSupplierContactNameLength+1)) }, ErrSupplierContactNameTooLong},
		{"long phone", func(f *SupplierFields) { f.Phone = ptr(strings.Repeat("1", MaxSupplierPhoneLength+1)) }, ErrSupplierPhoneTooLong},
		{"email without domain", func(f *SupplierFields) { f.Email = ptr("pedidos@") }, ErrSupplierEmailInvalid},
		{"long email", func(f *SupplierFields) { f.Email = ptr(strings.Repeat("a", MaxEmailLength) + "@x.com") }, ErrSupplierEmailInvalid},
		{"long address", func(f *SupplierFields) { f.Address = ptr(strings.Repeat("a", MaxSupplierAddressLength+1)) }, ErrSupplierAddressTooLong},
		{"long notes", func(f *SupplierFields) { f.Notes = ptr(strings.Repeat("a", MaxSupplierNotesLength+1)) }, ErrSupplierNotesTooLong},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fields := validSupplierFields()
			test.change(&fields)

			if _, err := NewSupplier(fields, supplierTime); !errors.Is(err, test.want) {
				t.Fatalf("NewSupplier() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestSupplierReplace(t *testing.T) {
	supplier, _ := NewSupplier(validSupplierFields(), supplierTime)
	supplier.ID = "s1"
	later := supplierTime.Add(time.Hour)

	replaced, err := supplier.Replace(SupplierFields{BusinessName: "Distribuidora del Centro"}, later)
	if err != nil {
		t.Fatalf("Replace() error = %v", err)
	}
	if replaced.ID != "s1" || replaced.BusinessName != "Distribuidora del Centro" || replaced.CUIT != nil || replaced.Notes != nil ||
		!replaced.CreatedAt.Equal(supplierTime) || !replaced.UpdatedAt.Equal(later) {
		t.Fatalf("replaced = %+v", replaced)
	}

	if _, err := supplier.Replace(SupplierFields{}, later); !errors.Is(err, ErrSupplierBusinessNameRequired) {
		t.Fatalf("Replace(empty) error = %v", err)
	}
}
