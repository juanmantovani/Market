package domain

import (
	"strings"
	"time"
)

const (
	MaxSupplierBusinessNameLength = 120
	MaxSupplierContactNameLength  = 120
	MaxSupplierPhoneLength        = 30
	MaxSupplierAddressLength      = 200
	MaxSupplierNotesLength        = 500
)

var (
	ErrSupplierNotFound             = &Error{Kind: KindNotFound, Code: "supplier_not_found", Message: "El proveedor no existe"}
	ErrSupplierBusinessNameRequired = &Error{Kind: KindInvalid, Code: "supplier_business_name_required", Message: "La razón social es obligatoria"}
	ErrSupplierBusinessNameTooLong  = &Error{Kind: KindInvalid, Code: "supplier_business_name_too_long", Message: "La razón social admite hasta 120 caracteres"}
	ErrSupplierCUITInvalid          = &Error{Kind: KindInvalid, Code: "supplier_cuit_invalid", Message: "El CUIT lleva 11 dígitos"}
	ErrSupplierCUITInUse            = &Error{Kind: KindConflict, Code: "supplier_cuit_in_use", Message: "Ya hay un proveedor con ese CUIT"}
	ErrSupplierInUse                = &Error{Kind: KindConflict, Code: "supplier_in_use", Message: "El proveedor está asignado a artículos: quitarlo de esos artículos antes de eliminarlo"}
	ErrSupplierContactNameTooLong   = &Error{Kind: KindInvalid, Code: "supplier_contact_name_too_long", Message: "El nombre de contacto admite hasta 120 caracteres"}
	ErrSupplierEmailInvalid         = &Error{Kind: KindInvalid, Code: "supplier_email_invalid", Message: "El email no es válido"}
	ErrSupplierPhoneTooLong         = &Error{Kind: KindInvalid, Code: "supplier_phone_too_long", Message: "El teléfono admite hasta 30 caracteres"}
	ErrSupplierAddressTooLong       = &Error{Kind: KindInvalid, Code: "supplier_address_too_long", Message: "La dirección admite hasta 200 caracteres"}
	ErrSupplierNotesTooLong         = &Error{Kind: KindInvalid, Code: "supplier_notes_too_long", Message: "Las notas admiten hasta 500 caracteres"}
)

// SupplierFields are the editable fields of a supplier. Creating and editing
// both take the whole set, so an optional field is cleared by leaving it
// empty.
type SupplierFields struct {
	BusinessName string
	CUIT         *string
	ContactName  *string
	Phone        *string
	Email        *string
	Address      *string
	Notes        *string
}

type Supplier struct {
	ID string
	SupplierFields
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewSupplier validates the fields of a supplier about to be created. The
// repository assigns its ID.
func NewSupplier(fields SupplierFields, now time.Time) (Supplier, error) {
	valid, err := validateSupplierFields(fields)
	if err != nil {
		return Supplier{}, err
	}

	return Supplier{SupplierFields: valid, CreatedAt: now, UpdatedAt: now}, nil
}

// Replace returns the supplier with all its editable fields replaced.
func (supplier Supplier) Replace(fields SupplierFields, now time.Time) (Supplier, error) {
	valid, err := validateSupplierFields(fields)
	if err != nil {
		return Supplier{}, err
	}

	updated := supplier
	updated.SupplierFields = valid
	updated.UpdatedAt = now

	return updated, nil
}

func validateSupplierFields(fields SupplierFields) (SupplierFields, error) {
	var err error
	valid := fields

	if valid.BusinessName, err = requiredText(fields.BusinessName, MaxSupplierBusinessNameLength, ErrSupplierBusinessNameRequired, ErrSupplierBusinessNameTooLong); err != nil {
		return SupplierFields{}, err
	}
	if valid.CUIT, err = validateSupplierCUIT(fields.CUIT); err != nil {
		return SupplierFields{}, err
	}
	if valid.ContactName, err = optionalText(fields.ContactName, MaxSupplierContactNameLength, ErrSupplierContactNameTooLong); err != nil {
		return SupplierFields{}, err
	}
	if valid.Phone, err = optionalText(fields.Phone, MaxSupplierPhoneLength, ErrSupplierPhoneTooLong); err != nil {
		return SupplierFields{}, err
	}
	if valid.Email, err = validateEmail(fields.Email, ErrSupplierEmailInvalid); err != nil {
		return SupplierFields{}, err
	}
	if valid.Address, err = optionalText(fields.Address, MaxSupplierAddressLength, ErrSupplierAddressTooLong); err != nil {
		return SupplierFields{}, err
	}
	if valid.Notes, err = optionalText(fields.Notes, MaxSupplierNotesLength, ErrSupplierNotesTooLong); err != nil {
		return SupplierFields{}, err
	}

	return valid, nil
}

// validateSupplierCUIT keeps only the digits, like a customer's document, so
// the same CUIT written with or without dashes is detected as repeated.
func validateSupplierCUIT(cuit *string) (*string, error) {
	if cuit == nil {
		return nil, nil
	}
	digits := documentSeparators.Replace(strings.TrimSpace(*cuit))
	if digits == "" {
		return nil, nil
	}
	if !digitsOnly.MatchString(digits) || len(digits) != 11 {
		return nil, ErrSupplierCUITInvalid
	}

	return &digits, nil
}
