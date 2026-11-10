package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var categoryTime = time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)

func ptr[T any](value T) *T {
	return &value
}

func TestNewCategory(t *testing.T) {
	description := "  Gaseosas, jugos y aguas  "

	category, err := NewCategory("  Bebidas ", &description, categoryTime)
	if err != nil {
		t.Fatalf("NewCategory() error = %v", err)
	}

	if category.Name != "Bebidas" {
		t.Errorf("Name = %q, want trimmed Bebidas", category.Name)
	}
	if category.Description == nil || *category.Description != "Gaseosas, jugos y aguas" {
		t.Errorf("Description = %v, want trimmed text", category.Description)
	}
	if category.Description == &description {
		t.Error("Description aliases the caller's pointer")
	}
	if !category.Active || !category.CreatedAt.Equal(categoryTime) || !category.UpdatedAt.Equal(categoryTime) {
		t.Errorf("category = %+v, want active with both timestamps set", category)
	}
	if category.ID != "" {
		t.Errorf("ID = %q, want empty until the repository assigns it", category.ID)
	}
}

func TestNewCategoryTreatsBlankDescriptionAsAbsent(t *testing.T) {
	category, err := NewCategory("Almacén", ptr("   "), categoryTime)
	if err != nil {
		t.Fatalf("NewCategory() error = %v", err)
	}
	if category.Description != nil {
		t.Fatalf("Description = %q, want nil", *category.Description)
	}
}

func TestNewCategoryValidates(t *testing.T) {
	tests := []struct {
		name        string
		inputName   string
		description *string
		want        *Error
	}{
		{name: "missing name", inputName: "   ", want: ErrCategoryNameRequired},
		{name: "long name", inputName: strings.Repeat("a", MaxCategoryNameLength+1), want: ErrCategoryNameTooLong},
		{name: "long description", inputName: "Bebidas", description: ptr(strings.Repeat("é", MaxCategoryDescriptionLength+1)), want: ErrCategoryDescriptionTooLong},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewCategory(test.inputName, test.description, categoryTime); !errors.Is(err, test.want) {
				t.Fatalf("NewCategory() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestNewCategoryCountsCharactersNotBytes(t *testing.T) {
	if _, err := NewCategory(strings.Repeat("ñ", MaxCategoryNameLength), nil, categoryTime); err != nil {
		t.Fatalf("NewCategory() with %d multibyte runes error = %v", MaxCategoryNameLength, err)
	}
}

func TestCategoryApply(t *testing.T) {
	original, _ := NewCategory("Bebidas", ptr("Gaseosas"), categoryTime)
	later := categoryTime.Add(time.Hour)

	renamed, err := original.Apply(CategoryChanges{Name: ptr(" Bebidas sin alcohol ")}, later)
	if err != nil {
		t.Fatalf("Apply(name) error = %v", err)
	}
	if renamed.Name != "Bebidas sin alcohol" || renamed.Description == nil || *renamed.Description != "Gaseosas" {
		t.Fatalf("Apply(name) = %+v, want new name and untouched description", renamed)
	}
	if !renamed.UpdatedAt.Equal(later) || !renamed.CreatedAt.Equal(categoryTime) {
		t.Fatalf("timestamps = %v/%v, want only UpdatedAt moved", renamed.CreatedAt, renamed.UpdatedAt)
	}

	cleared, err := original.Apply(CategoryChanges{Description: ptr("")}, later)
	if err != nil {
		t.Fatalf("Apply(clear description) error = %v", err)
	}
	if cleared.Description != nil || cleared.Name != "Bebidas" {
		t.Fatalf("Apply(clear description) = %+v", cleared)
	}
	if original.Description == nil {
		t.Fatal("Apply mutated the original category")
	}
}

func TestCategoryApplyRejectsInvalidChanges(t *testing.T) {
	original, _ := NewCategory("Bebidas", nil, categoryTime)

	tests := []struct {
		name    string
		changes CategoryChanges
		want    *Error
	}{
		{name: "empty", changes: CategoryChanges{}, want: ErrCategoryUpdateEmpty},
		{name: "blank name", changes: CategoryChanges{Name: ptr(" ")}, want: ErrCategoryNameRequired},
		{name: "long description", changes: CategoryChanges{Description: ptr(strings.Repeat("x", 201))}, want: ErrCategoryDescriptionTooLong},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := original.Apply(test.changes, categoryTime); !errors.Is(err, test.want) {
				t.Fatalf("Apply() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestCategoryWithActive(t *testing.T) {
	original, _ := NewCategory("Bebidas", nil, categoryTime)
	later := categoryTime.Add(time.Minute)

	inactive := original.WithActive(false, later)

	if inactive.Active || !inactive.UpdatedAt.Equal(later) {
		t.Fatalf("WithActive(false) = %+v", inactive)
	}
	if !original.Active {
		t.Fatal("WithActive mutated the original category")
	}
}

func TestCategoryNameKey(t *testing.T) {
	category := Category{Name: "  Lácteos  y Fiambres"}

	if got := category.NameKey(); got != "lacteos y fiambres" {
		t.Fatalf("NameKey() = %q", got)
	}
}
