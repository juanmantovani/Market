package domain

import "time"

const (
	MaxCategoryNameLength        = 60
	MaxCategoryDescriptionLength = 200
)

var (
	ErrCategoryNotFound           = &Error{Kind: KindNotFound, Code: "category_not_found", Message: "La categoría no existe"}
	ErrCategoryNameRequired       = &Error{Kind: KindInvalid, Code: "category_name_required", Message: "El nombre es obligatorio"}
	ErrCategoryNameTooLong        = &Error{Kind: KindInvalid, Code: "category_name_too_long", Message: "El nombre admite hasta 60 caracteres"}
	ErrCategoryDescriptionTooLong = &Error{Kind: KindInvalid, Code: "category_description_too_long", Message: "La descripción admite hasta 200 caracteres"}
	ErrCategoryNameInUse          = &Error{Kind: KindConflict, Code: "category_name_in_use", Message: "Ya hay una categoría con ese nombre"}
	ErrCategoryUpdateEmpty        = &Error{Kind: KindInvalid, Code: "category_update_empty", Message: "No se enviaron cambios"}
)

// Category groups items so they are easier to find and filter. Its name is
// unique, ignoring case, accents and spacing.
type Category struct {
	ID          string
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewCategory validates the fields of a category about to be created. The
// repository assigns its ID.
func NewCategory(name string, description *string, now time.Time) (Category, error) {
	validName, err := validateCategoryName(name)
	if err != nil {
		return Category{}, err
	}
	validDescription, err := optionalText(description, MaxCategoryDescriptionLength, ErrCategoryDescriptionTooLong)
	if err != nil {
		return Category{}, err
	}

	return Category{
		Name:        validName,
		Description: validDescription,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// CategoryChanges is a partial update: a nil field stays as it is, and a
// blank description clears it.
type CategoryChanges struct {
	Name        *string
	Description *string
}

func (changes CategoryChanges) IsEmpty() bool {
	return changes.Name == nil && changes.Description == nil
}

// Apply returns the category with changes applied.
func (category Category) Apply(changes CategoryChanges, now time.Time) (Category, error) {
	if changes.IsEmpty() {
		return Category{}, ErrCategoryUpdateEmpty
	}

	updated := category
	if changes.Name != nil {
		name, err := validateCategoryName(*changes.Name)
		if err != nil {
			return Category{}, err
		}
		updated.Name = name
	}
	if changes.Description != nil {
		description, err := optionalText(changes.Description, MaxCategoryDescriptionLength, ErrCategoryDescriptionTooLong)
		if err != nil {
			return Category{}, err
		}
		updated.Description = description
	}
	updated.UpdatedAt = now

	return updated, nil
}

// NameKey is the form of the name used to detect duplicates.
func (category Category) NameKey() string {
	return NormalizeText(category.Name)
}

func validateCategoryName(name string) (string, error) {
	return requiredText(name, MaxCategoryNameLength, ErrCategoryNameRequired, ErrCategoryNameTooLong)
}
