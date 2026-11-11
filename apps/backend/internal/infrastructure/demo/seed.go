// Package demo loads sample records so a fresh in-memory backend has
// something to show. It goes through the use cases, so demo data obeys the
// same rules as data entered from the app.
package demo

import (
	"context"
	"fmt"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/categories"
	"market/backend/internal/business/usecase/items"
)

var seeder = domain.Actor{ID: "demo", Name: "Datos demo", Role: domain.RoleAdmin}

type Catalog struct {
	CreateCategory categories.CreateCategory
	CreateItem     items.CreateItem
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

type demoItem struct {
	category string
	code     string
	barcode  string
	name     string
	unit     domain.Unit
	price    domain.Money
	cost     domain.Money
	taxRate  domain.Percent
	minStock domain.Quantity
}

var demoItems = []demoItem{
	{"Almacén", "ALM-001", "7790000000011", "Yerba mate 1 kg", domain.UnitPiece, 450_000, 310_000, 2100, domain.WholeQuantity(6)},
	{"Almacén", "ALM-002", "7790000000028", "Arroz largo fino 1 kg", domain.UnitPiece, 180_000, 120_000, 2100, domain.WholeQuantity(10)},
	{"Almacén", "ALM-003", "7790000000035", "Fideos tirabuzón 500 g", domain.UnitPiece, 120_000, 80_000, 2100, domain.WholeQuantity(12)},
	{"Almacén", "ALM-004", "7790000000042", "Aceite de girasol 1,5 l", domain.UnitPiece, 390_000, 280_000, 2100, domain.WholeQuantity(6)},
	{"Bebidas", "BEB-001", "7790000000059", "Gaseosa cola 2,25 l", domain.UnitPiece, 320_000, 220_000, 2100, domain.WholeQuantity(12)},
	{"Bebidas", "BEB-002", "7790000000066", "Agua mineral 1,5 l", domain.UnitPiece, 110_000, 70_000, 2100, domain.WholeQuantity(24)},
	{"Lácteos", "LAC-001", "7790000000073", "Leche entera 1 l", domain.UnitPiece, 150_000, 105_000, 1050, domain.WholeQuantity(20)},
	{"Lácteos", "LAC-002", "7790000000080", "Yogur bebible frutilla 1 l", domain.UnitPiece, 210_000, 150_000, 2100, domain.WholeQuantity(8)},
	{"Panadería", "PAN-001", "", "Pan francés", domain.UnitKilogram, 280_000, 160_000, 1050, domain.WholeQuantity(5)},
	{"Panadería", "PAN-002", "", "Facturas", domain.UnitPiece, 60_000, 30_000, 1050, 0},
	{"Fiambrería", "FIA-001", "", "Queso cremoso", domain.UnitKilogram, 980_000, 690_000, 2100, 2500},
	{"Fiambrería", "FIA-002", "", "Jamón cocido", domain.UnitKilogram, 1_250_000, 880_000, 2100, 2000},
	{"Limpieza", "LIM-001", "7790000000097", "Lavandina 1 l", domain.UnitPiece, 95_000, 60_000, 2100, domain.WholeQuantity(10)},
	{"Perfumería", "PER-001", "7790000000103", "Jabón de tocador", domain.UnitPiece, 85_000, 50_000, 2100, domain.WholeQuantity(10)},
	{"Kiosco", "KIO-001", "7790000000110", "Alfajor triple", domain.UnitPiece, 110_000, 70_000, 2100, domain.WholeQuantity(20)},
}

func Seed(ctx context.Context, catalog Catalog) error {
	categoryIDs := map[string]string{}
	for _, category := range demoCategories {
		description := category.description
		created, err := catalog.CreateCategory.Execute(ctx, categories.CreateCategoryInput{
			Actor:       seeder,
			Name:        category.name,
			Description: &description,
		})
		if err != nil {
			return fmt.Errorf("seeding category %q: %w", category.name, err)
		}
		categoryIDs[category.name] = created.ID
	}

	for _, item := range demoItems {
		barcode, cost := item.barcode, item.cost
		_, err := catalog.CreateItem.Execute(ctx, items.CreateItemInput{
			Actor: seeder,
			Fields: domain.ItemFields{
				Code:        item.code,
				Barcode:     &barcode,
				Name:        item.name,
				CategoryID:  categoryIDs[item.category],
				Unit:        item.unit,
				SalePrice:   item.price,
				Cost:        &cost,
				TaxRate:     item.taxRate,
				TracksStock: true,
				MinStock:    item.minStock,
			},
		})
		if err != nil {
			return fmt.Errorf("seeding item %q: %w", item.code, err)
		}
	}

	return nil
}
