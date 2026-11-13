package sales

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway/gatewayfake"
)

var (
	admin   = domain.Actor{ID: "u-admin", Role: domain.RoleAdmin}
	cashier = domain.Actor{ID: "u-cashier", Role: domain.RoleCashier}

	errDown = errors.New("connection refused")
)

func ptr[T any](value T) *T {
	return &value
}

func assertError(t *testing.T, err error, want *domain.Error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %s", err, want.Code)
	}
}

type fakes struct {
	tx        *gatewayfake.TxManager
	sales     *gatewayfake.SaleRepository
	items     *gatewayfake.ItemRepository
	customers *gatewayfake.CustomerRepository
	sellers   *gatewayfake.SellerRepository
	stock     *gatewayfake.StockRepository
	policy    domain.StockPolicy
}

func newFakes() *fakes {
	stocked := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	return &fakes{
		tx:    &gatewayfake.TxManager{},
		sales: &gatewayfake.SaleRepository{},
		items: &gatewayfake.ItemRepository{Stored: map[string]domain.Item{
			"coca":  {ID: "coca", ItemFields: domain.ItemFields{Code: "BEB-1", Name: "Gaseosa cola", Unit: domain.UnitPiece, SalePrice: 320_000, TaxRate: 2100}},
			"queso": {ID: "queso", ItemFields: domain.ItemFields{Code: "FIA-1", Name: "Queso cremoso", Unit: domain.UnitKilogram, SalePrice: 980_000, TaxRate: 2100}},
		}},
		customers: &gatewayfake.CustomerRepository{GetResult: domain.Customer{ID: "c1", CustomerFields: domain.CustomerFields{
			Name: "Ana Pérez", DocumentType: domain.DocumentDNI, DocumentNumber: "28456789", TaxCondition: domain.TaxFinalConsumer,
		}}},
		sellers: &gatewayfake.SellerRepository{GetResult: domain.Seller{ID: "s1", SellerFields: domain.SellerFields{Name: "Juan Pérez", Code: "JP"}}},
		stock: &gatewayfake.StockRepository{Levels: map[string]domain.StockLevel{
			"coca":  {ItemID: "coca", Quantity: domain.WholeQuantity(18), UpdatedAt: stocked},
			"queso": {ItemID: "queso", Quantity: 1_800, UpdatedAt: stocked},
		}},
	}
}

func (fakes *fakes) useCase() CreateSale {
	return NewCreateSale(fakes.tx, fakes.sales, fakes.items, fakes.customers, fakes.sellers, fakes.stock, fakes.policy)
}

// request sells 2 colas and 0,250 kg of cheese with 10 % off:
// 640.000 + 245.000 − 24.500 = 860.500 cents.
func request() domain.SaleRequest {
	return domain.SaleRequest{
		IdempotencyKey: "key-1",
		CustomerID:     ptr("c1"),
		SellerID:       ptr("s1"),
		Lines: []domain.SaleRequestLine{
			{ItemID: "coca", Quantity: domain.WholeQuantity(2)},
			{ItemID: "queso", Quantity: 250, Discount: &domain.Discount{Type: domain.DiscountPercent, Value: 1_000}},
		},
		Payments:      []domain.Payment{{Method: domain.PaymentDebitCard, Amount: 500_000}, {Method: domain.PaymentCash, Amount: 400_000}},
		ExpectedTotal: 860_500,
	}
}

func TestCreateSaleIsAllowedToCashiers(t *testing.T) {
	fakes := newFakes()
	before := time.Now().UTC()

	sale, err := fakes.useCase().Execute(context.Background(), CreateSaleInput{Actor: cashier, Request: request()})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if sale.ID != "sale-1" || sale.Number != 1 || sale.Total != 860_500 || sale.DiscountTotal != 24_500 || sale.Change != 39_500 ||
		sale.Customer.Name != "Ana Pérez" || sale.Seller == nil || sale.Seller.Code != "JP" || sale.ActorID != "u-cashier" ||
		sale.Status != domain.SaleCompleted || sale.CreatedAt.Before(before) || sale.IdempotencyKey != "key-1" {
		t.Fatalf("sale = %+v", sale)
	}
	if fakes.tx.Calls != 1 || len(fakes.sales.Created) != 1 {
		t.Fatalf("units of work = %d, sales created = %d; want 1 and 1", fakes.tx.Calls, len(fakes.sales.Created))
	}
	if len(fakes.customers.GetIDs) != 1 || fakes.customers.GetIDs[0] != "c1" || len(fakes.sellers.GetIDs) != 1 || fakes.sellers.GetIDs[0] != "s1" {
		t.Fatalf("read customers %v and sellers %v", fakes.customers.GetIDs, fakes.sellers.GetIDs)
	}

	recorded := fakes.stock.Recorded
	if len(recorded) != 2 {
		t.Fatalf("recorded %d movements, want one per line", len(recorded))
	}
	coca, queso := recorded[0], recorded[1]
	if coca.ItemID != "coca" || coca.Type != domain.MovementSale || coca.Quantity != -2_000 || coca.BalanceAfter != 16_000 ||
		*coca.Reference != "Venta 1" || coca.ActorID != "u-cashier" || !coca.CreatedAt.Equal(sale.CreatedAt) {
		t.Fatalf("coca movement = %+v", coca)
	}
	if queso.ItemID != "queso" || queso.Quantity != -250 || queso.BalanceAfter != 1_550 {
		t.Fatalf("cheese movement = %+v", queso)
	}
}

func TestCreateSaleWithoutCustomerGoesToFinalConsumer(t *testing.T) {
	fakes := newFakes()
	fakes.customers.GetResult = domain.FinalConsumer(time.Now().UTC())
	withoutParties := request()
	withoutParties.CustomerID = nil
	withoutParties.SellerID = nil

	sale, err := fakes.useCase().Execute(context.Background(), CreateSaleInput{Actor: admin, Request: withoutParties})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if fakes.customers.GetIDs[0] != domain.FinalConsumerID || sale.Customer.ID != domain.FinalConsumerID || sale.Customer.Name != "Consumidor Final" {
		t.Fatalf("customer read %v, sale customer %+v", fakes.customers.GetIDs, sale.Customer)
	}
	if len(fakes.sellers.GetIDs) != 0 || sale.Seller != nil {
		t.Fatalf("a sale without seller read sellers %v and got %+v", fakes.sellers.GetIDs, sale.Seller)
	}
}

func TestCreateSaleReplaysARetry(t *testing.T) {
	fakes := newFakes()
	useCase := fakes.useCase()
	first, err := useCase.Execute(context.Background(), CreateSaleInput{Actor: cashier, Request: request()})
	if err != nil {
		t.Fatalf("first Execute() error = %v", err)
	}

	retry, err := useCase.Execute(context.Background(), CreateSaleInput{Actor: cashier, Request: request()})
	if err != nil {
		t.Fatalf("retried Execute() error = %v", err)
	}

	if retry.ID != first.ID || retry.Number != first.Number || len(fakes.sales.Created) != 1 || len(fakes.stock.Recorded) != 2 {
		t.Fatalf("retry = %+v; %d sales and %d movements, want the first sale and nothing new", retry, len(fakes.sales.Created), len(fakes.stock.Recorded))
	}

	changed := request()
	changed.Lines[0].Quantity = domain.WholeQuantity(3)
	_, err = useCase.Execute(context.Background(), CreateSaleInput{Actor: cashier, Request: changed})
	assertError(t, err, domain.ErrIdempotencyKeyConflict)
	if len(fakes.sales.Created) != 1 || len(fakes.stock.Recorded) != 2 {
		t.Fatal("a conflicting key created a sale or moved stock")
	}
}

func TestCreateSaleKeyBelongsToTheActor(t *testing.T) {
	fakes := newFakes()
	useCase := fakes.useCase()

	first, err := useCase.Execute(context.Background(), CreateSaleInput{Actor: cashier, Request: request()})
	if err != nil {
		t.Fatalf("first Execute() error = %v", err)
	}
	other, err := useCase.Execute(context.Background(), CreateSaleInput{Actor: admin, Request: request()})
	if err != nil {
		t.Fatalf("Execute() with another actor's key error = %v", err)
	}

	if other.ID == first.ID || other.ActorID != admin.ID || len(fakes.sales.Created) != 2 {
		t.Fatalf("other actor got %+v; %d sales, want their own new sale", other, len(fakes.sales.Created))
	}
}

func TestCreateSaleWithInsufficientStock(t *testing.T) {
	fakes := newFakes()
	tooMuch := request()
	tooMuch.Lines[1].Quantity = 2_000
	tooMuch.ExpectedTotal = 640_000 + 1_960_000 - 196_000
	tooMuch.Payments = []domain.Payment{{Method: domain.PaymentCash, Amount: tooMuch.ExpectedTotal}}

	_, err := fakes.useCase().Execute(context.Background(), CreateSaleInput{Actor: cashier, Request: tooMuch})

	assertError(t, err, domain.ErrInsufficientStock)
	if !strings.Contains(err.Error(), "Queso cremoso") {
		t.Fatalf("message = %q, want the item name", err.Error())
	}

	fakes = newFakes()
	fakes.policy = domain.StockPolicy{AllowNegative: true}
	if _, err := fakes.useCase().Execute(context.Background(), CreateSaleInput{Actor: cashier, Request: tooMuch}); err != nil {
		t.Fatalf("Execute(negative stock allowed) error = %v", err)
	}
	if balance := fakes.stock.Levels["queso"].Quantity; balance != -200 {
		t.Fatalf("cheese stock = %d, want -200", balance)
	}
}

func TestCreateSaleErrors(t *testing.T) {
	tests := []struct {
		name    string
		actor   *domain.Actor
		change  func(*fakes, *domain.SaleRequest)
		want    *domain.Error
		noUnits bool
	}{
		{name: "no actor", actor: &domain.Actor{}, want: domain.ErrForbidden, noUnits: true},
		{name: "invalid request", change: func(_ *fakes, r *domain.SaleRequest) { r.Lines = nil }, want: domain.ErrSaleLinesRequired, noUnits: true},
		{name: "unit of work unavailable", change: func(f *fakes, _ *domain.SaleRequest) { f.tx.Err = errDown }, want: domain.ErrStorageUnavailable},
		{name: "idempotency lookup fails", change: func(f *fakes, _ *domain.SaleRequest) { f.sales.GetByKeyErr = errDown }, want: domain.ErrStorageUnavailable},
		{name: "unknown customer", change: func(f *fakes, _ *domain.SaleRequest) { f.customers.GetErr = domain.ErrCustomerNotFound }, want: domain.ErrSaleCustomerNotFound},
		{name: "customers unavailable", change: func(f *fakes, _ *domain.SaleRequest) { f.customers.GetErr = errDown }, want: domain.ErrStorageUnavailable},
		{name: "unknown seller", change: func(f *fakes, _ *domain.SaleRequest) { f.sellers.GetErr = domain.ErrSellerNotFound }, want: domain.ErrSaleSellerNotFound},
		{name: "unknown item", change: func(_ *fakes, r *domain.SaleRequest) { r.Lines[0].ItemID = "agua" }, want: domain.ErrSaleItemNotFound},
		{name: "items unavailable", change: func(f *fakes, _ *domain.SaleRequest) { f.items.GetErr = errDown }, want: domain.ErrStorageUnavailable},
		{name: "total mismatch", change: func(_ *fakes, r *domain.SaleRequest) { r.ExpectedTotal = 860_000 }, want: domain.ErrSaleTotalMismatch},
		{name: "key taken meanwhile", change: func(f *fakes, _ *domain.SaleRequest) { f.sales.CreateErr = domain.ErrIdempotencyKeyConflict }, want: domain.ErrIdempotencyKeyConflict},
		{name: "stock unavailable", change: func(f *fakes, _ *domain.SaleRequest) { f.stock.GetLevelErr = errDown }, want: domain.ErrStorageUnavailable},
		{name: "movement not recorded", change: func(f *fakes, _ *domain.SaleRequest) { f.stock.RecordErr = errDown }, want: domain.ErrStorageUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakes, saleRequest := newFakes(), request()
			if test.change != nil {
				test.change(fakes, &saleRequest)
			}
			actor := cashier
			if test.actor != nil {
				actor = *test.actor
			}

			_, err := fakes.useCase().Execute(context.Background(), CreateSaleInput{Actor: actor, Request: saleRequest})

			assertError(t, err, test.want)
			if test.noUnits && fakes.tx.Calls != 0 {
				t.Fatal("a rejected request opened a unit of work")
			}
		})
	}
}
