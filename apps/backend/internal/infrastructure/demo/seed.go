// Package demo loads sample records so a fresh in-memory backend has
// something to show. It goes through the use cases, so demo data obeys the
// same rules as data entered from the app.
package demo

import (
	"context"
	"fmt"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/usecase/categories"
	"market/backend/internal/business/usecase/customers"
	"market/backend/internal/business/usecase/items"
	"market/backend/internal/business/usecase/sellers"
	"market/backend/internal/business/usecase/stock"
	"market/backend/internal/business/usecase/suppliers"
)

var seeder = domain.Actor{ID: "demo", Name: "Datos demo", Role: domain.RoleAdmin}

// UseCases are the use cases the seed loads data through.
type UseCases struct {
	CreateCategory categories.CreateCategory
	CreateItem     items.CreateItem
	CreateCustomer customers.CreateCustomer
	CreateSupplier suppliers.CreateSupplier
	CreateSeller   sellers.CreateSeller

	RegisterInitialStock    stock.RegisterInitialStock
	RegisterStockEntry      stock.RegisterStockEntry
	RegisterStockAdjustment stock.RegisterStockAdjustment
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

type demoCustomer struct {
	name         string
	documentType domain.DocumentType
	document     string
	taxCondition domain.TaxCondition
	email        string
	phone        string
}

var demoCustomers = []demoCustomer{
	{"Ana Pérez", domain.DocumentDNI, "28456789", domain.TaxFinalConsumer, "", "351 555-0101"},
	{"Carlos Gómez", domain.DocumentDNI, "30111222", domain.TaxFinalConsumer, "carlos.gomez@example.com", ""},
	{"Kiosco La Esquina SRL", domain.DocumentCUIT, "30712345678", domain.TaxRegistered, "compras@laesquina.example.com", "351 555-0303"},
	{"María López", domain.DocumentCUIT, "27234567894", domain.TaxMonotax, "maria.lopez@example.com", "351 555-0404"},
	{"Juan (cliente del barrio)", domain.DocumentNone, "", domain.TaxFinalConsumer, "", ""},
}

type demoSupplier struct {
	businessName string
	cuit         string
	contactName  string
	phone        string
	email        string
}

var demoSuppliers = []demoSupplier{
	{"Distribuidora del Centro SA", "30701234561", "Laura Díaz", "351 555-2000", "pedidos@delcentro.example.com"},
	{"Lácteos del Sur SRL", "30709876543", "Martín Ruiz", "351 555-2100", ""},
	{"Mayorista Norte SA", "30715555559", "", "", "ventas@mayoristanorte.example.com"},
	{"Panificadora La Espiga", "", "Rosa", "351 555-2200", ""},
}

type demoSeller struct {
	name       string
	code       string
	commission *domain.Percent
}

func commission(rate domain.Percent) *domain.Percent {
	return &rate
}

var demoSellers = []demoSeller{
	{"Juan Pérez", "JP", commission(250)},
	{"Lucía Fernández", "LF", commission(300)},
	{"María Gómez", "MG", nil},
}

// demoItemSuppliers says who supplies the items of each category; the
// fiambrería has none, to show an item without suppliers.
var demoItemSuppliers = map[string][]string{
	"Almacén":    {"Distribuidora del Centro SA", "Mayorista Norte SA"},
	"Bebidas":    {"Mayorista Norte SA"},
	"Lácteos":    {"Lácteos del Sur SRL"},
	"Panadería":  {"Panificadora La Espiga"},
	"Limpieza":   {"Mayorista Norte SA"},
	"Perfumería": {"Mayorista Norte SA"},
	"Kiosco":     {"Distribuidora del Centro SA"},
}

// demoInitialStock is the opening stock of each item, in thousandths. The
// yerba, the milk and the cremoso cheese start below their minimum, so the
// low stock filter has something to show.
var demoInitialStock = map[string]domain.Quantity{
	"ALM-001": domain.WholeQuantity(4),
	"ALM-002": domain.WholeQuantity(25),
	"ALM-003": domain.WholeQuantity(40),
	"ALM-004": domain.WholeQuantity(15),
	"BEB-001": domain.WholeQuantity(18),
	"BEB-002": domain.WholeQuantity(48),
	"LAC-001": domain.WholeQuantity(12),
	"LAC-002": domain.WholeQuantity(10),
	"PAN-001": 8_500,
	"PAN-002": domain.WholeQuantity(36),
	"FIA-001": 1_800,
	"FIA-002": 3_200,
	"LIM-001": domain.WholeQuantity(18),
	"PER-001": domain.WholeQuantity(22),
	"KIO-001": domain.WholeQuantity(60),
}

func Seed(ctx context.Context, useCases UseCases) error {
	categoryIDs := map[string]string{}
	for _, category := range demoCategories {
		description := category.description
		created, err := useCases.CreateCategory.Execute(ctx, categories.CreateCategoryInput{
			Actor:       seeder,
			Name:        category.name,
			Description: &description,
		})
		if err != nil {
			return fmt.Errorf("seeding category %q: %w", category.name, err)
		}
		categoryIDs[category.name] = created.ID
	}

	supplierIDs := map[string]string{}
	for _, supplier := range demoSuppliers {
		cuit, contactName, phone, email := supplier.cuit, supplier.contactName, supplier.phone, supplier.email
		created, err := useCases.CreateSupplier.Execute(ctx, suppliers.CreateSupplierInput{
			Actor: seeder,
			Fields: domain.SupplierFields{
				BusinessName: supplier.businessName,
				CUIT:         &cuit,
				ContactName:  &contactName,
				Phone:        &phone,
				Email:        &email,
			},
		})
		if err != nil {
			return fmt.Errorf("seeding supplier %q: %w", supplier.businessName, err)
		}
		supplierIDs[supplier.businessName] = created.ID
	}

	itemIDs := map[string]string{}
	for _, item := range demoItems {
		var itemSupplierIDs []string
		for _, name := range demoItemSuppliers[item.category] {
			itemSupplierIDs = append(itemSupplierIDs, supplierIDs[name])
		}
		barcode, cost := item.barcode, item.cost
		created, err := useCases.CreateItem.Execute(ctx, items.CreateItemInput{
			Actor: seeder,
			Fields: domain.ItemFields{
				Code:        item.code,
				Barcode:     &barcode,
				Name:        item.name,
				CategoryID:  categoryIDs[item.category],
				SupplierIDs: itemSupplierIDs,
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
		itemIDs[item.code] = created.ID
	}

	if err := seedStock(ctx, useCases, itemIDs, supplierIDs); err != nil {
		return err
	}

	for _, customer := range demoCustomers {
		email, phone := customer.email, customer.phone
		_, err := useCases.CreateCustomer.Execute(ctx, customers.CreateCustomerInput{
			Actor: seeder,
			Fields: domain.CustomerFields{
				Name:           customer.name,
				DocumentType:   customer.documentType,
				DocumentNumber: customer.document,
				TaxCondition:   customer.taxCondition,
				Email:          &email,
				Phone:          &phone,
			},
		})
		if err != nil {
			return fmt.Errorf("seeding customer %q: %w", customer.name, err)
		}
	}

	for _, seller := range demoSellers {
		_, err := useCases.CreateSeller.Execute(ctx, sellers.CreateSellerInput{
			Actor:  seeder,
			Fields: domain.SellerFields{Name: seller.name, Code: seller.code, CommissionRate: seller.commission},
		})
		if err != nil {
			return fmt.Errorf("seeding seller %q: %w", seller.code, err)
		}
	}

	return nil
}

func seedStock(ctx context.Context, useCases UseCases, itemIDs, supplierIDs map[string]string) error {
	for _, item := range demoItems {
		_, err := useCases.RegisterInitialStock.Execute(ctx, stock.RegisterInitialStockInput{
			Actor: seeder, ItemID: itemIDs[item.code], Quantity: demoInitialStock[item.code],
		})
		if err != nil {
			return fmt.Errorf("seeding the stock of %q: %w", item.code, err)
		}
	}

	supplierID, reference, cost := supplierIDs["Mayorista Norte SA"], "Remito 0001-00004512", domain.Money(220_000)
	_, err := useCases.RegisterStockEntry.Execute(ctx, stock.RegisterStockEntryInput{
		Actor: seeder,
		Entry: domain.StockEntry{
			SupplierID: &supplierID,
			Reference:  &reference,
			Lines:      []domain.StockEntryLine{{ItemID: itemIDs["BEB-001"], Quantity: domain.WholeQuantity(12), UnitCost: &cost}},
		},
	})
	if err != nil {
		return fmt.Errorf("seeding a stock entry: %w", err)
	}

	_, err = useCases.RegisterStockAdjustment.Execute(ctx, stock.RegisterStockAdjustmentInput{
		Actor:      seeder,
		Adjustment: domain.StockAdjustment{ItemID: itemIDs["KIO-001"], Quantity: domain.WholeQuantity(-2), Reason: "Rotura en góndola"},
	})
	if err != nil {
		return fmt.Errorf("seeding a stock adjustment: %w", err)
	}

	return nil
}
