package domain

import (
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaxCustomerNameLength    = 120
	MaxCustomerEmailLength   = 254
	MaxCustomerPhoneLength   = 30
	MaxCustomerAddressLength = 200
	MaxCustomerNotesLength   = 500
)

// DocumentType is the identity document a customer is registered with.
type DocumentType string

const (
	DocumentNone DocumentType = "none"
	DocumentDNI  DocumentType = "dni"
	DocumentCUIT DocumentType = "cuit"
	DocumentCUIL DocumentType = "cuil"
)

// TaxCondition is the customer's condition before IVA (AFIP/ARCA).
type TaxCondition string

const (
	TaxFinalConsumer TaxCondition = "final_consumer"
	TaxMonotax       TaxCondition = "monotax"
	TaxRegistered    TaxCondition = "registered"
	TaxExempt        TaxCondition = "exempt"
)

var (
	ErrCustomerNotFound                 = &Error{Kind: KindNotFound, Code: "customer_not_found", Message: "El cliente no existe"}
	ErrCustomerNameRequired             = &Error{Kind: KindInvalid, Code: "customer_name_required", Message: "El nombre es obligatorio"}
	ErrCustomerNameTooLong              = &Error{Kind: KindInvalid, Code: "customer_name_too_long", Message: "El nombre admite hasta 120 caracteres"}
	ErrCustomerDocumentTypeInvalid      = &Error{Kind: KindInvalid, Code: "customer_document_type_invalid", Message: "El tipo de documento no es válido"}
	ErrCustomerDocumentInvalid          = &Error{Kind: KindInvalid, Code: "customer_document_invalid", Message: "El número de documento no es válido: el DNI lleva 7 u 8 dígitos y el CUIT o CUIL, 11"}
	ErrCustomerDocumentInUse            = &Error{Kind: KindConflict, Code: "customer_document_in_use", Message: "Ya hay un cliente con ese documento"}
	ErrCustomerTaxConditionInvalid      = &Error{Kind: KindInvalid, Code: "customer_tax_condition_invalid", Message: "La condición frente al IVA no es válida"}
	ErrCustomerTaxConditionRequiresCUIT = &Error{Kind: KindInvalid, Code: "customer_tax_condition_requires_cuit", Message: "Para monotributo, responsable inscripto o exento hace falta un CUIT"}
	ErrCustomerEmailInvalid             = &Error{Kind: KindInvalid, Code: "customer_email_invalid", Message: "El email no es válido"}
	ErrCustomerPhoneTooLong             = &Error{Kind: KindInvalid, Code: "customer_phone_too_long", Message: "El teléfono admite hasta 30 caracteres"}
	ErrCustomerAddressTooLong           = &Error{Kind: KindInvalid, Code: "customer_address_too_long", Message: "La dirección admite hasta 200 caracteres"}
	ErrCustomerNotesTooLong             = &Error{Kind: KindInvalid, Code: "customer_notes_too_long", Message: "Las notas admiten hasta 500 caracteres"}
	documentSeparators                  = strings.NewReplacer(" ", "", ".", "", "-", "", "/", "")
	digitsOnly                          = regexp.MustCompile(`^[0-9]+$`)
	emailShape                          = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
)

// CustomerFields are the editable fields of a customer. Creating and editing
// both take the whole set, so an optional field is cleared by leaving it
// empty.
type CustomerFields struct {
	Name           string
	DocumentType   DocumentType
	DocumentNumber string
	TaxCondition   TaxCondition
	Email          *string
	Phone          *string
	Address        *string
	Notes          *string
}

type Customer struct {
	ID string
	CustomerFields
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewCustomer validates the fields of a customer about to be created. The
// repository assigns its ID.
func NewCustomer(fields CustomerFields, now time.Time) (Customer, error) {
	valid, err := validateCustomerFields(fields)
	if err != nil {
		return Customer{}, err
	}

	return Customer{CustomerFields: valid, CreatedAt: now, UpdatedAt: now}, nil
}

// Replace returns the customer with all its editable fields replaced.
func (customer Customer) Replace(fields CustomerFields, now time.Time) (Customer, error) {
	valid, err := validateCustomerFields(fields)
	if err != nil {
		return Customer{}, err
	}

	updated := customer
	updated.CustomerFields = valid
	updated.UpdatedAt = now

	return updated, nil
}

// DocumentKey identifies the customer's document for duplicate checks. It is
// empty for a customer without document, who never collides.
func (customer Customer) DocumentKey() string {
	if customer.DocumentType == DocumentNone {
		return ""
	}

	return string(customer.DocumentType) + ":" + customer.DocumentNumber
}

func validateCustomerFields(fields CustomerFields) (CustomerFields, error) {
	var err error
	valid := fields

	if valid.Name, err = requiredText(fields.Name, MaxCustomerNameLength, ErrCustomerNameRequired, ErrCustomerNameTooLong); err != nil {
		return CustomerFields{}, err
	}
	if valid.DocumentNumber, err = validateDocument(fields.DocumentType, fields.DocumentNumber); err != nil {
		return CustomerFields{}, err
	}
	if err = validateTaxCondition(fields.TaxCondition, fields.DocumentType); err != nil {
		return CustomerFields{}, err
	}
	if valid.Email, err = validateEmail(fields.Email); err != nil {
		return CustomerFields{}, err
	}
	if valid.Phone, err = optionalText(fields.Phone, MaxCustomerPhoneLength, ErrCustomerPhoneTooLong); err != nil {
		return CustomerFields{}, err
	}
	if valid.Address, err = optionalText(fields.Address, MaxCustomerAddressLength, ErrCustomerAddressTooLong); err != nil {
		return CustomerFields{}, err
	}
	if valid.Notes, err = optionalText(fields.Notes, MaxCustomerNotesLength, ErrCustomerNotesTooLong); err != nil {
		return CustomerFields{}, err
	}

	return valid, nil
}

// validateDocument keeps only the digits of the number, so "20-12345678-9"
// and "20123456789" are the same CUIT. Check digits aren't verified: the aim
// is to catch typos in length, not to validate against ARCA.
func validateDocument(documentType DocumentType, number string) (string, error) {
	digits := documentSeparators.Replace(strings.TrimSpace(number))

	switch documentType {
	case DocumentNone:
		if digits != "" {
			return "", ErrCustomerDocumentInvalid
		}
		return "", nil
	case DocumentDNI:
		if !digitsOnly.MatchString(digits) || len(digits) < 7 || len(digits) > 8 {
			return "", ErrCustomerDocumentInvalid
		}
	case DocumentCUIT, DocumentCUIL:
		if !digitsOnly.MatchString(digits) || len(digits) != 11 {
			return "", ErrCustomerDocumentInvalid
		}
	default:
		return "", ErrCustomerDocumentTypeInvalid
	}

	return digits, nil
}

func validateTaxCondition(condition TaxCondition, documentType DocumentType) error {
	switch condition {
	case TaxFinalConsumer:
		return nil
	case TaxMonotax, TaxRegistered, TaxExempt:
		if documentType != DocumentCUIT {
			return ErrCustomerTaxConditionRequiresCUIT
		}
		return nil
	default:
		return ErrCustomerTaxConditionInvalid
	}
}

func validateEmail(email *string) (*string, error) {
	if email == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*email)
	if trimmed == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(trimmed) > MaxCustomerEmailLength || !emailShape.MatchString(trimmed) {
		return nil, ErrCustomerEmailInvalid
	}

	return &trimmed, nil
}
