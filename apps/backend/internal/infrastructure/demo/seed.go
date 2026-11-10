// Package demo loads sample records so a fresh in-memory backend has
// something to show. It goes through the use cases, so demo data obeys the
// same rules as data entered from the app.
package demo

import (
	"context"
	"fmt"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/categories"
)

var seeder = domain.Actor{ID: "demo", Name: "Datos demo", Role: domain.RoleAdmin}

type Catalog struct {
	CreateCategory categories.CreateCategory
}

type demoCategory struct {
	name        string
	description string
}

var demoCategories = []demoCategory{
	{"Almacén", "Fideos, arroz, harinas, aceites y conservas"},
	{"Bebidas", "Gaseosas, aguas, jugos y bebidas con alcohol"},
	{"Lácteos", "Leches, yogures, quesos y manteca"},
	{"Panadería", "Pan, facturas y galletitas"},
	{"Fiambrería", "Fiambres y quesos al corte"},
	{"Limpieza", "Artículos de limpieza para el hogar"},
	{"Perfumería", "Higiene y cuidado personal"},
	{"Kiosco", "Golosinas, snacks y cigarrillos"},
}

func Seed(ctx context.Context, catalog Catalog) error {
	for _, category := range demoCategories {
		description := category.description
		_, err := catalog.CreateCategory.Execute(ctx, categories.CreateCategoryInput{
			Actor:       seeder,
			Name:        category.name,
			Description: &description,
		})
		if err != nil {
			return fmt.Errorf("seeding category %q: %w", category.name, err)
		}
	}

	return nil
}
