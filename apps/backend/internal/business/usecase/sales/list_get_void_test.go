package sales

import (
	"context"
	"testing"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
	"market/backend/internal/business/gateway/gatewayfake"
)

var manager = domain.Actor{ID: "u-manager", Role: domain.RoleManager}

func storedSale() domain.Sale {
	return domain.Sale{
		ID: "sale-1", Number: 7, Status: domain.SaleCompleted, ActorID: "u-cashier",
		Customer: domain.SaleCustomer{ID: domain.FinalConsumerID, Name: "Consumidor Final"},
		Lines: []domain.SaleLine{
			{ItemID: "coca", Name: "Gaseosa cola", Unit: domain.UnitPiece, Quantity: domain.WholeQuantity(2), UnitPrice: 320_000, LineTotal: 640_000},
			{ItemID: "queso", Name: "Queso cremoso", Unit: domain.UnitKilogram, Quantity: 250, UnitPrice: 980_000, LineTotal: 245_000},
		},
		Subtotal: 885_000, Total: 885_000,
		CreatedAt: time.Date(2026, 11, 12, 10, 0, 0, 0, time.UTC),
	}
}

func TestListSalesIsAllowedToCashiers(t *testing.T) {
	page := domain.Page[domain.Sale]{Items: []domain.Sale{storedSale()}, Total: 1, Page: 2, Limit: 10}
	repository := &gatewayfake.SaleRepository{ListResult: page}
	from := time.Date(2026, 11, 12, 3, 0, 0, 0, time.UTC)

	got, err := NewListSales(repository).Execute(context.Background(), ListSalesInput{
		Actor:  cashier,
		Filter: domain.SaleFilter{From: from, To: from.Add(24 * time.Hour), CustomerID: " c1 ", SellerID: "s1", Status: domain.SaleCompleted},
		Page:   2,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got.Total != 1 || got.Items[0].ID != "sale-1" {
		t.Fatalf("page = %+v", got)
	}
	want := gateway.SaleFilter{
		SaleFilter: domain.SaleFilter{From: from, To: from.Add(24 * time.Hour), CustomerID: "c1", SellerID: "s1", Status: domain.SaleCompleted},
		Page:       domain.PageRequest{Page: 2, Limit: 10},
	}
	if len(repository.ListFilters) != 1 || repository.ListFilters[0] != want {
		t.Fatalf("filters = %+v, want %+v", repository.ListFilters, want)
	}
}

func TestListSalesErrors(t *testing.T) {
	tests := []struct {
		name  string
		input ListSalesInput
		fake  *gatewayfake.SaleRepository
		want  *domain.Error
	}{
		{name: "no actor", input: ListSalesInput{}, want: domain.ErrForbidden},
		{name: "unknown status", input: ListSalesInput{Actor: cashier, Filter: domain.SaleFilter{Status: "pending"}}, want: domain.ErrSaleStatusInvalid},
		{name: "invalid page", input: ListSalesInput{Actor: cashier, Limit: 101}, want: domain.ErrInvalidPage},
		{name: "storage down", input: ListSalesInput{Actor: cashier}, fake: &gatewayfake.SaleRepository{ListErr: errDown}, want: domain.ErrStorageUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := test.fake
			if repository == nil {
				repository = &gatewayfake.SaleRepository{}
			}

			_, err := NewListSales(repository).Execute(context.Background(), test.input)

			assertError(t, err, test.want)
		})
	}
}

func TestGetSale(t *testing.T) {
	repository := &gatewayfake.SaleRepository{GetResult: storedSale()}

	got, err := NewGetSale(repository).Execute(context.Background(), GetSaleInput{Actor: cashier, ID: "sale-1"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got.Number != 7 || repository.GetIDs[0] != "sale-1" {
		t.Fatalf("sale = %+v, read %v", got, repository.GetIDs)
	}

	_, err = NewGetSale(repository).Execute(context.Background(), GetSaleInput{ID: "sale-1"})
	assertError(t, err, domain.ErrForbidden)
	_, err = NewGetSale(&gatewayfake.SaleRepository{GetErr: domain.ErrSaleNotFound}).Execute(context.Background(), GetSaleInput{Actor: cashier, ID: "x"})
	assertError(t, err, domain.ErrSaleNotFound)
	_, err = NewGetSale(&gatewayfake.SaleRepository{GetErr: errDown}).Execute(context.Background(), GetSaleInput{Actor: cashier, ID: "x"})
	assertError(t, err, domain.ErrStorageUnavailable)
}

type voidFakes struct {
	tx    *gatewayfake.TxManager
	sales *gatewayfake.SaleRepository
	stock *gatewayfake.StockRepository
}

func newVoidFakes() voidFakes {
	return voidFakes{
		tx:    &gatewayfake.TxManager{},
		sales: &gatewayfake.SaleRepository{GetResult: storedSale()},
		stock: &gatewayfake.StockRepository{Levels: map[string]domain.StockLevel{
			"coca":  {ItemID: "coca", Quantity: domain.WholeQuantity(16)},
			"queso": {ItemID: "queso", Quantity: -50},
		}},
	}
}

func (fakes voidFakes) useCase() VoidSale {
	return NewVoidSale(fakes.tx, fakes.sales, fakes.stock)
}

func TestVoidSale(t *testing.T) {
	fakes := newVoidFakes()
	before := time.Now().UTC()

	voided, err := fakes.useCase().Execute(context.Background(), VoidSaleInput{Actor: manager, ID: "sale-1", Reason: " Error de carga "})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if voided.Status != domain.SaleVoided || *voided.VoidReason != "Error de carga" || *voided.VoidedBy != "u-manager" || voided.VoidedAt.Before(before) {
		t.Fatalf("voided = %+v", voided)
	}
	if fakes.tx.Calls != 1 || len(fakes.sales.Voided) != 1 || fakes.sales.GetIDs[0] != "sale-1" {
		t.Fatalf("units of work %d, voided %d, read %v", fakes.tx.Calls, len(fakes.sales.Voided), fakes.sales.GetIDs)
	}
	recorded := fakes.stock.Recorded
	if len(recorded) != 2 {
		t.Fatalf("recorded %d movements, want one per line", len(recorded))
	}
	coca, queso := recorded[0], recorded[1]
	if coca.Type != domain.MovementSaleVoid || coca.Quantity != 2_000 || coca.BalanceAfter != 18_000 || *coca.Reference != "Venta 7" ||
		coca.ActorID != "u-manager" || !coca.CreatedAt.Equal(*voided.VoidedAt) {
		t.Fatalf("coca movement = %+v", coca)
	}
	if queso.ItemID != "queso" || queso.Quantity != 250 || queso.BalanceAfter != 200 {
		t.Fatalf("cheese movement = %+v", queso)
	}
}

func TestVoidSaleErrors(t *testing.T) {
	voided := storedSale()
	voided.Status = domain.SaleVoided

	tests := []struct {
		name    string
		actor   domain.Actor
		reason  string
		change  func(voidFakes)
		want    *domain.Error
		noUnits bool
	}{
		{name: "cashiers cannot void", actor: cashier, reason: "Error", want: domain.ErrForbidden, noUnits: true},
		{name: "unit of work unavailable", actor: manager, reason: "Error", change: func(f voidFakes) { f.tx.Err = errDown }, want: domain.ErrStorageUnavailable},
		{name: "unknown sale", actor: manager, reason: "Error", change: func(f voidFakes) { f.sales.GetErr = domain.ErrSaleNotFound }, want: domain.ErrSaleNotFound},
		{name: "already voided", actor: admin, reason: "Error", change: func(f voidFakes) { f.sales.GetResult = voided }, want: domain.ErrSaleAlreadyVoided},
		{name: "no reason", actor: manager, reason: " ", want: domain.ErrSaleVoidReasonRequired},
		{name: "void not saved", actor: manager, reason: "Error", change: func(f voidFakes) { f.sales.VoidErr = errDown }, want: domain.ErrStorageUnavailable},
		{name: "stock unavailable", actor: manager, reason: "Error", change: func(f voidFakes) { f.stock.GetLevelErr = errDown }, want: domain.ErrStorageUnavailable},
		{name: "balance out of range", actor: manager, reason: "Error", change: func(f voidFakes) {
			f.stock.Levels["coca"] = domain.StockLevel{ItemID: "coca", Quantity: domain.MaxStockBalance}
		}, want: domain.ErrStockBalanceOutOfRange},
		{name: "movement not recorded", actor: manager, reason: "Error", change: func(f voidFakes) { f.stock.RecordErr = errDown }, want: domain.ErrStorageUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakes := newVoidFakes()
			if test.change != nil {
				test.change(fakes)
			}

			_, err := fakes.useCase().Execute(context.Background(), VoidSaleInput{Actor: test.actor, ID: "sale-1", Reason: test.reason})

			assertError(t, err, test.want)
			if test.noUnits && fakes.tx.Calls != 0 {
				t.Fatal("a forbidden void opened a unit of work")
			}
		})
	}
}
