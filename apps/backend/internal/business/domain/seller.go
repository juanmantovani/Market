package domain

import (
	"regexp"
	"strings"
	"time"
)

const (
	MaxSellerNameLength = 80
	MaxSellerCodeLength = 6
)

var (
	ErrSellerNotFound          = &Error{Kind: KindNotFound, Code: "seller_not_found", Message: "El vendedor no existe"}
	ErrSellerNameRequired      = &Error{Kind: KindInvalid, Code: "seller_name_required", Message: "El nombre es obligatorio"}
	ErrSellerNameTooLong       = &Error{Kind: KindInvalid, Code: "seller_name_too_long", Message: "El nombre admite hasta 80 caracteres"}
	ErrSellerCodeRequired      = &Error{Kind: KindInvalid, Code: "seller_code_required", Message: "El código es obligatorio"}
	ErrSellerCodeInvalid       = &Error{Kind: KindInvalid, Code: "seller_code_invalid", Message: "El código lleva de 1 a 6 letras o números, sin espacios"}
	ErrSellerCodeInUse         = &Error{Kind: KindConflict, Code: "seller_code_in_use", Message: "Ya hay un vendedor con ese código"}
	ErrSellerCommissionInvalid = &Error{Kind: KindInvalid, Code: "seller_commission_invalid", Message: "La comisión va de 0 a 100 %"}
	ErrSellerInUse             = &Error{Kind: KindConflict, Code: "seller_in_use", Message: "El vendedor tiene ventas registradas y no se puede eliminar"}
	sellerCodeShape            = regexp.MustCompile(`^[A-Z0-9]{1,6}$`)
)

// SellerFields are the editable fields of a seller. A seller is who attends
// a sale, not a user of the app.
type SellerFields struct {
	Name string
	// Code is what the cashier types to pick the seller at the counter. It
	// is stored in upper case, so "jp" and "JP" are the same code.
	Code string
	// CommissionRate is informative until commissions are reported.
	CommissionRate *Percent
}

type Seller struct {
	ID string
	SellerFields
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewSeller validates the fields of a seller about to be created. The
// repository assigns its ID.
func NewSeller(fields SellerFields, now time.Time) (Seller, error) {
	valid, err := validateSellerFields(fields)
	if err != nil {
		return Seller{}, err
	}

	return Seller{SellerFields: valid, CreatedAt: now, UpdatedAt: now}, nil
}

// Replace returns the seller with all its editable fields replaced.
func (seller Seller) Replace(fields SellerFields, now time.Time) (Seller, error) {
	valid, err := validateSellerFields(fields)
	if err != nil {
		return Seller{}, err
	}

	updated := seller
	updated.SellerFields = valid
	updated.UpdatedAt = now

	return updated, nil
}

func validateSellerFields(fields SellerFields) (SellerFields, error) {
	var err error
	valid := fields

	if valid.Name, err = requiredText(fields.Name, MaxSellerNameLength, ErrSellerNameRequired, ErrSellerNameTooLong); err != nil {
		return SellerFields{}, err
	}

	valid.Code = strings.ToUpper(strings.TrimSpace(fields.Code))
	if valid.Code == "" {
		return SellerFields{}, ErrSellerCodeRequired
	}
	if !sellerCodeShape.MatchString(valid.Code) {
		return SellerFields{}, ErrSellerCodeInvalid
	}

	if fields.CommissionRate != nil {
		if !fields.CommissionRate.IsValid() {
			return SellerFields{}, ErrSellerCommissionInvalid
		}
		rate := *fields.CommissionRate
		valid.CommissionRate = &rate
	}

	return valid, nil
}
