package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var customerTime = time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)

func validCustomerFields() CustomerFields {
	return CustomerFields{
		Name:           "  Kiosco La Esquina SRL ",
		DocumentType:   DocumentCUIT,
		DocumentNumber: " 30-71234567-8 ",
		TaxCondition:   TaxRegistered,
		Email:          ptr(" compras@laesquina.com.ar "),
		Phone:          ptr(" 351 555-1234 "),
		Address:        ptr(" San Martín 123, Córdoba "),
		Notes:          ptr(" Paga a 30 días "),
	}
}

func TestNewCustomer(t *testing.T) {
	customer, err := NewCustomer(validCustomerFields(), customerTime)
	if err != nil {
		t.Fatalf("NewCustomer() error = %v", err)
	}

	if customer.Name != "Kiosco La Esquina SRL" || customer.DocumentNumber != "30712345678" {
		t.Errorf("customer = %+v", customer.CustomerFields)
	}
	if *customer.Email != "compras@laesquina.com.ar" || *customer.Phone != "351 555-1234" ||
		*customer.Address != "San Martín 123, Córdoba" || *customer.Notes != "Paga a 30 días" {
		t.Errorf("optional fields not trimmed: %+v", customer.CustomerFields)
	}
	if !customer.CreatedAt.Equal(customerTime) || !customer.UpdatedAt.Equal(customerTime) || customer.ID != "" {
		t.Errorf("customer = %+v", customer)
	}
	if customer.DocumentKey() != "cuit:30712345678" {
		t.Errorf("DocumentKey() = %q", customer.DocumentKey())
	}
}

func TestNewCustomerWithoutDocumentOrContact(t *testing.T) {
	customer, err := NewCustomer(CustomerFields{
		Name:         "Juan",
		DocumentType: DocumentNone,
		TaxCondition: TaxFinalConsumer,
		Email:        ptr("  "),
	}, customerTime)
	if err != nil {
		t.Fatalf("NewCustomer() error = %v", err)
	}

	if customer.DocumentNumber != "" || customer.DocumentKey() != "" || customer.Email != nil || customer.Phone != nil {
		t.Fatalf("customer = %+v", customer.CustomerFields)
	}
}

func TestNewCustomerAcceptsDNIAndCUIL(t *testing.T) {
	for _, fields := range []CustomerFields{
		{Name: "Ana", DocumentType: DocumentDNI, DocumentNumber: "12.345.678", TaxCondition: TaxFinalConsumer},
		{Name: "Beto", DocumentType: DocumentDNI, DocumentNumber: "9876543", TaxCondition: TaxFinalConsumer},
		{Name: "Carla", DocumentType: DocumentCUIL, DocumentNumber: "27-12345678-4", TaxCondition: TaxFinalConsumer},
		{Name: "Dario", DocumentType: DocumentCUIT, DocumentNumber: "20123456789", TaxCondition: TaxMonotax},
		{Name: "Eva", DocumentType: DocumentCUIT, DocumentNumber: "20123456789", TaxCondition: TaxExempt},
	} {
		if _, err := NewCustomer(fields, customerTime); err != nil {
			t.Errorf("NewCustomer(%+v) error = %v", fields, err)
		}
	}
}

func TestNewCustomerValidates(t *testing.T) {
	tests := []struct {
		name   string
		change func(*CustomerFields)
		want   *Error
	}{
		{"missing name", func(f *CustomerFields) { f.Name = " " }, ErrCustomerNameRequired},
		{"long name", func(f *CustomerFields) { f.Name = strings.Repeat("n", MaxCustomerNameLength+1) }, ErrCustomerNameTooLong},
		{"unknown document type", func(f *CustomerFields) { f.DocumentType = "passport" }, ErrCustomerDocumentTypeInvalid},
		{"short CUIT", func(f *CustomerFields) { f.DocumentNumber = "30-7123456-8" }, ErrCustomerDocumentInvalid},
		{"CUIT with letters", func(f *CustomerFields) { f.DocumentNumber = "30-7123456A-8" }, ErrCustomerDocumentInvalid},
		{"short DNI", func(f *CustomerFields) {
			f.DocumentType, f.DocumentNumber, f.TaxCondition = DocumentDNI, "123456", TaxFinalConsumer
		}, ErrCustomerDocumentInvalid},
		{"long DNI", func(f *CustomerFields) {
			f.DocumentType, f.DocumentNumber, f.TaxCondition = DocumentDNI, "123456789", TaxFinalConsumer
		}, ErrCustomerDocumentInvalid},
		{"number without document type", func(f *CustomerFields) {
			f.DocumentType, f.TaxCondition = DocumentNone, TaxFinalConsumer
		}, ErrCustomerDocumentInvalid},
		{"unknown tax condition", func(f *CustomerFields) { f.TaxCondition = "vip" }, ErrCustomerTaxConditionInvalid},
		{"registered without CUIT", func(f *CustomerFields) {
			f.DocumentType, f.DocumentNumber = DocumentDNI, "12345678"
		}, ErrCustomerTaxConditionRequiresCUIT},
		{"monotax with CUIL", func(f *CustomerFields) {
			f.DocumentType, f.TaxCondition = DocumentCUIL, TaxMonotax
		}, ErrCustomerTaxConditionRequiresCUIT},
		{"email without domain", func(f *CustomerFields) { f.Email = ptr("compras@") }, ErrCustomerEmailInvalid},
		{"email with spaces", func(f *CustomerFields) { f.Email = ptr("com pras@x.com") }, ErrCustomerEmailInvalid},
		{"long email", func(f *CustomerFields) { f.Email = ptr(strings.Repeat("a", 250) + "@x.com") }, ErrCustomerEmailInvalid},
		{"long phone", func(f *CustomerFields) { f.Phone = ptr(strings.Repeat("1", MaxCustomerPhoneLength+1)) }, ErrCustomerPhoneTooLong},
		{"long address", func(f *CustomerFields) { f.Address = ptr(strings.Repeat("a", MaxCustomerAddressLength+1)) }, ErrCustomerAddressTooLong},
		{"long notes", func(f *CustomerFields) { f.Notes = ptr(strings.Repeat("a", MaxCustomerNotesLength+1)) }, ErrCustomerNotesTooLong},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fields := validCustomerFields()
			test.change(&fields)

			if _, err := NewCustomer(fields, customerTime); !errors.Is(err, test.want) {
				t.Fatalf("NewCustomer() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestCustomerReplace(t *testing.T) {
	original, _ := NewCustomer(validCustomerFields(), customerTime)
	original.ID = "customer-1"
	later := customerTime.Add(time.Hour)

	fields := validCustomerFields()
	fields.Name = "La Esquina SA"
	fields.Notes = nil
	replaced, err := original.Replace(fields, later)
	if err != nil {
		t.Fatalf("Replace() error = %v", err)
	}

	if replaced.ID != "customer-1" || replaced.Name != "La Esquina SA" || replaced.Notes != nil ||
		!replaced.CreatedAt.Equal(customerTime) || !replaced.UpdatedAt.Equal(later) {
		t.Fatalf("replaced = %+v", replaced)
	}
	if original.Notes == nil {
		t.Fatal("Replace mutated the original customer")
	}

	fields.Name = ""
	if _, err := original.Replace(fields, later); !errors.Is(err, ErrCustomerNameRequired) {
		t.Fatalf("Replace(invalid) error = %v", err)
	}
}

func TestFinalConsumerIsASystemCustomer(t *testing.T) {
	finalConsumer := FinalConsumer(customerTime)

	if finalConsumer.ID != FinalConsumerID || finalConsumer.Name != "Consumidor Final" || !finalConsumer.System ||
		finalConsumer.DocumentType != DocumentNone || finalConsumer.DocumentKey() != "" || finalConsumer.TaxCondition != TaxFinalConsumer ||
		!finalConsumer.CreatedAt.Equal(customerTime) || !finalConsumer.UpdatedAt.Equal(customerTime) {
		t.Fatalf("FinalConsumer() = %+v", finalConsumer)
	}
	if _, err := NewCustomer(finalConsumer.CustomerFields, customerTime); err != nil {
		t.Fatalf("Consumidor Final has invalid fields: %v", err)
	}

	if !errors.Is(finalConsumer.EnsureChangeable(), ErrCustomerIsSystem) {
		t.Fatal("EnsureChangeable() accepted the system customer")
	}
	if _, err := finalConsumer.Replace(validCustomerFields(), customerTime); !errors.Is(err, ErrCustomerIsSystem) {
		t.Fatalf("Replace(system customer) error = %v, want ErrCustomerIsSystem", err)
	}

	customer, _ := NewCustomer(validCustomerFields(), customerTime)
	if customer.System || customer.EnsureChangeable() != nil {
		t.Fatalf("a new customer is a system one: %+v", customer)
	}
}
