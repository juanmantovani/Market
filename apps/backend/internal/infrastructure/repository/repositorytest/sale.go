package repositorytest

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

// saleParties are the stored records a contract sale refers to.
type saleParties struct {
	customer domain.Customer
	seller   domain.Seller
	coca     domain.Item
	pan      domain.Item
}

func newSaleParties(t *testing.T, repositories Repositories) saleParties {
	t.Helper()
	bebidas := mustCreate(t, repositories.Categories, newCategory(t, "Bebidas", nil))
	pan := newItem(t, bebidas.ID, "PAN-1", "Pan francés")
	pan.Unit = domain.UnitKilogram
	return saleParties{
		customer: mustCreateCustomer(t, repositories.Customers, newCustomer(t, "Ana Pérez", domain.DocumentDNI, "28456789")),
		seller:   mustCreateSeller(t, repositories.Sellers, newSeller(t, "Juan Pérez", "JP")),
		coca:     mustCreateItem(t, repositories.Items, newItem(t, bebidas.ID, "BEB-1", "Gaseosa cola")),
		pan:      mustCreateItem(t, repositories.Items, pan),
	}
}

// newSale is a completed sale of 2 colas with $ 200 off and 0,750 kg of
// bread with 10 % off, plus a 10 % general discount, paid by debit card and
// cash, made under key.
func newSale(parties saleParties, key string) domain.Sale {
	return domain.Sale{
		Customer: domain.SaleCustomer{
			ID: parties.customer.ID, Name: parties.customer.Name, DocumentType: parties.customer.DocumentType,
			DocumentNumber: parties.customer.DocumentNumber, TaxCondition: parties.customer.TaxCondition,
		},
		Seller: &domain.SaleSeller{ID: parties.seller.ID, Name: parties.seller.Name, Code: parties.seller.Code},
		Lines: []domain.SaleLine{
			{
				ItemID: parties.coca.ID, Code: "BEB-1", Name: "Gaseosa cola", Unit: domain.UnitPiece, TaxRate: 2100, Quantity: 2_000, UnitPrice: 100_000,
				Discount: &domain.Discount{Type: domain.DiscountAmount, Value: 20_000}, DiscountAmount: 20_000, CartDiscountAmount: 18_000, LineTotal: 162_000,
			},
			{
				ItemID: parties.pan.ID, Code: "PAN-1", Name: "Pan francés", Unit: domain.UnitKilogram, TaxRate: 2100, Quantity: 750, UnitPrice: 100_000,
				Discount: &domain.Discount{Type: domain.DiscountPercent, Value: 1_000}, DiscountAmount: 7_500, CartDiscountAmount: 6_750, LineTotal: 60_750,
			},
		},
		Subtotal:           275_000,
		CartDiscount:       &domain.Discount{Type: domain.DiscountPercent, Value: 1_000},
		CartDiscountAmount: 24_750,
		DiscountTotal:      52_250,
		Total:              222_750,
		Payments: []domain.Payment{
			{Method: domain.PaymentDebitCard, Amount: 200_000, Reference: text("Cupón 4471")},
			{Method: domain.PaymentCash, Amount: 100_000},
		},
		Change:         77_250,
		Status:         domain.SaleCompleted,
		ActorID:        "u-contract",
		IdempotencyKey: key,
		Fingerprint:    "fingerprint-" + key,
		CreatedAt:      contractTime,
	}
}

func mustCreateSale(t *testing.T, repository gateway.SaleRepository, sale domain.Sale) domain.Sale {
	t.Helper()
	created, err := repository.Create(context.Background(), sale)
	if err != nil {
		t.Fatalf("Create(sale %q) error = %v", sale.IdempotencyKey, err)
	}
	return created
}

// describeSale prints every field of a sale, dereferencing pointers and
// with the time in UTC, so two sales compare by value.
func describeSale(sale domain.Sale) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "%s #%d %s customer=%+v", sale.ID, sale.Number, sale.Status, sale.Customer)
	if sale.Seller != nil {
		fmt.Fprintf(&builder, " seller=%+v", *sale.Seller)
	}
	for _, line := range sale.Lines {
		fmt.Fprintf(&builder, " line=%+v", line)
	}
	for _, payment := range sale.Payments {
		fmt.Fprintf(&builder, " payment=%s:%d:%s", payment.Method, payment.Amount, valueOf(payment.Reference))
	}
	fmt.Fprintf(&builder, " subtotal=%d cart-discount=%s:%d discount=%d total=%d change=%d by %s key=%s fingerprint=%s at %s",
		sale.Subtotal, sale.CartDiscount, sale.CartDiscountAmount, sale.DiscountTotal, sale.Total, sale.Change, sale.ActorID, sale.IdempotencyKey, sale.Fingerprint,
		sale.CreatedAt.UTC().Format(time.RFC3339Nano))
	if sale.VoidedAt != nil {
		fmt.Fprintf(&builder, " voided at %s", sale.VoidedAt.UTC().Format(time.RFC3339Nano))
	}
	fmt.Fprintf(&builder, " reason=%s by=%s", valueOf(sale.VoidReason), valueOf(sale.VoidedBy))

	return builder.String()
}

func saleNumbers(sales []domain.Sale) []int64 {
	result := make([]int64, len(sales))
	for i, sale := range sales {
		result[i] = sale.Number
	}
	return result
}

func valueOf(value *string) string {
	if value == nil {
		return "<nil>"
	}
	return *value
}

// SaleRepository runs the gateway.SaleRepository contract, including the
// records a sale refers to and how it shares the unit of work of the
// adapter's TxManager.
func SaleRepository(t *testing.T, newRepositories NewRepositories) {
	ctx := context.Background()

	t.Run("create assigns an id and the next number and stores every field", func(t *testing.T) {
		repositories := newRepositories(t)
		parties := newSaleParties(t, repositories)

		first := mustCreateSale(t, repositories.Sales, newSale(parties, "key-1"))
		second := mustCreateSale(t, repositories.Sales, newSale(parties, "key-2"))

		if first.ID == "" || second.ID == "" || first.ID == second.ID {
			t.Fatalf("Create() IDs = %q, %q", first.ID, second.ID)
		}
		if first.Number != 1 || second.Number != 2 {
			t.Fatalf("Create() numbers = %d, %d, want 1 and 2", first.Number, second.Number)
		}
		want := newSale(parties, "key-1")
		want.ID, want.Number = first.ID, 1
		if describeSale(first) != describeSale(want) {
			t.Fatalf("Create() = %s\nwant %s", describeSale(first), describeSale(want))
		}

		got, err := repositories.Sales.GetByIdempotencyKey(ctx, "u-contract", "key-1")
		if err != nil {
			t.Fatalf("GetByIdempotencyKey() error = %v", err)
		}
		if describeSale(got) != describeSale(want) {
			t.Fatalf("GetByIdempotencyKey() = %s\nwant %s", describeSale(got), describeSale(want))
		}
	})

	t.Run("a sale without seller or discounts", func(t *testing.T) {
		repositories := newRepositories(t)
		sale := newSale(newSaleParties(t, repositories), "key-1")
		sale.Seller, sale.CartDiscount = nil, nil
		for i := range sale.Lines {
			sale.Lines[i].Discount = nil
		}

		mustCreateSale(t, repositories.Sales, sale)

		got, err := repositories.Sales.GetByIdempotencyKey(ctx, "u-contract", "key-1")
		if err != nil || got.Seller != nil || got.CartDiscount != nil || got.Lines[0].Discount != nil || got.Lines[1].Discount != nil {
			t.Fatalf("GetByIdempotencyKey() = %s, %v; want no seller and no discounts", describeSale(got), err)
		}
	})

	t.Run("an unknown key is not found", func(t *testing.T) {
		repositories := newRepositories(t)
		mustCreateSale(t, repositories.Sales, newSale(newSaleParties(t, repositories), "key-1"))

		if _, err := repositories.Sales.GetByIdempotencyKey(ctx, "u-contract", "key-2"); !errors.Is(err, domain.ErrSaleNotFound) {
			t.Fatalf("GetByIdempotencyKey(unknown) error = %v, want ErrSaleNotFound", err)
		}
	})

	t.Run("an idempotency key makes one sale only", func(t *testing.T) {
		repositories := newRepositories(t)
		parties := newSaleParties(t, repositories)
		mustCreateSale(t, repositories.Sales, newSale(parties, "key-1"))

		if _, err := repositories.Sales.Create(ctx, newSale(parties, "key-1")); !errors.Is(err, domain.ErrIdempotencyKeyConflict) {
			t.Fatalf("Create(repeated key) error = %v, want ErrIdempotencyKeyConflict", err)
		}
		if next := mustCreateSale(t, repositories.Sales, newSale(parties, "key-2")); next.Number != 2 {
			t.Fatalf("number after a rejected sale = %d, want 2", next.Number)
		}
	})

	t.Run("an idempotency key belongs to the actor that used it", func(t *testing.T) {
		repositories := newRepositories(t)
		parties := newSaleParties(t, repositories)
		mine := mustCreateSale(t, repositories.Sales, newSale(parties, "key-1"))
		theirs := newSale(parties, "key-1")
		theirs.ActorID = "u-other"
		theirs = mustCreateSale(t, repositories.Sales, theirs)

		for actorID, want := range map[string]string{"u-contract": mine.ID, "u-other": theirs.ID} {
			got, err := repositories.Sales.GetByIdempotencyKey(ctx, actorID, "key-1")
			if err != nil || got.ID != want {
				t.Fatalf("GetByIdempotencyKey(%q) = %q, %v; want %q", actorID, got.ID, err, want)
			}
		}
		if _, err := repositories.Sales.GetByIdempotencyKey(ctx, "u-nobody", "key-1"); !errors.Is(err, domain.ErrSaleNotFound) {
			t.Fatalf("GetByIdempotencyKey(other actor) error = %v, want ErrSaleNotFound", err)
		}
	})

	t.Run("the customer, the seller and the items must exist", func(t *testing.T) {
		repositories := newRepositories(t)
		parties := newSaleParties(t, repositories)
		unknown := "00000000-0000-7000-8000-000000000000"
		tests := []struct {
			name   string
			change func(*domain.Sale)
			want   *domain.Error
		}{
			{"customer", func(sale *domain.Sale) { sale.Customer.ID = unknown }, domain.ErrSaleCustomerNotFound},
			{"seller", func(sale *domain.Sale) { sale.Seller.ID = unknown }, domain.ErrSaleSellerNotFound},
			{"item", func(sale *domain.Sale) { sale.Lines[1].ItemID = unknown }, domain.ErrSaleItemNotFound},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				sale := newSale(parties, "key-"+test.name)
				test.change(&sale)

				if _, err := repositories.Sales.Create(ctx, sale); !errors.Is(err, test.want) {
					t.Fatalf("Create(unknown %s) error = %v, want %s", test.name, err, test.want.Code)
				}
				if _, err := repositories.Sales.GetByIdempotencyKey(ctx, sale.ActorID, sale.IdempotencyKey); !errors.Is(err, domain.ErrSaleNotFound) {
					t.Fatalf("a rejected sale was stored: %v", err)
				}
			})
		}

		sold := newSale(parties, "key-final-consumer")
		sold.Customer = domain.SaleCustomer{ID: domain.FinalConsumerID, Name: "Consumidor Final", DocumentType: domain.DocumentNone, TaxCondition: domain.TaxFinalConsumer}
		mustCreateSale(t, repositories.Sales, sold)
	})

	t.Run("customers and sellers with sales cannot be deleted", func(t *testing.T) {
		repositories := newRepositories(t)
		parties := newSaleParties(t, repositories)
		mustCreateSale(t, repositories.Sales, newSale(parties, "key-1"))

		if err := repositories.Customers.Delete(ctx, parties.customer.ID); !errors.Is(err, domain.ErrCustomerInUse) {
			t.Fatalf("Customers.Delete(with sales) error = %v, want ErrCustomerInUse", err)
		}
		if err := repositories.Sellers.Delete(ctx, parties.seller.ID); !errors.Is(err, domain.ErrSellerInUse) {
			t.Fatalf("Sellers.Delete(with sales) error = %v, want ErrSellerInUse", err)
		}

		beto := mustCreateCustomer(t, repositories.Customers, newCustomer(t, "Beto", domain.DocumentNone, ""))
		lucia := mustCreateSeller(t, repositories.Sellers, newSeller(t, "Lucía", "LF"))
		if err := repositories.Customers.Delete(ctx, beto.ID); err != nil {
			t.Fatalf("Customers.Delete(without sales) error = %v", err)
		}
		if err := repositories.Sellers.Delete(ctx, lucia.ID); err != nil {
			t.Fatalf("Sellers.Delete(without sales) error = %v", err)
		}
	})

	t.Run("a failed unit of work saves neither the sale nor its movements", func(t *testing.T) {
		repositories := newRepositories(t)
		parties := newSaleParties(t, repositories)
		failure := errors.New("the stock ran short")

		err := repositories.Tx.WithinTx(ctx, func(ctx context.Context) error {
			sale, err := repositories.Sales.Create(ctx, newSale(parties, "key-1"))
			if err != nil {
				return err
			}
			movement := newMovement(parties.coca.ID, domain.MovementSale, -2_000, -2_000, sale.CreatedAt)
			movement.Reference = text(sale.Reference())
			if _, err := repositories.Stock.Record(ctx, movement); err != nil {
				return err
			}
			return failure
		})
		if !errors.Is(err, failure) {
			t.Fatalf("WithinTx() error = %v, want %v", err, failure)
		}

		if _, err := repositories.Sales.GetByIdempotencyKey(ctx, "u-contract", "key-1"); !errors.Is(err, domain.ErrSaleNotFound) {
			t.Fatalf("GetByIdempotencyKey(rolled back) error = %v, want ErrSaleNotFound", err)
		}
		movements, _ := repositories.Stock.ListMovements(ctx, gateway.StockMovementFilter{Page: domain.PageRequest{Page: 1, Limit: 10}})
		if movements.Total != 0 {
			t.Fatalf("movements after the rollback = %v", movementIDs(movements.Items))
		}
		if err := repositories.Sellers.Delete(ctx, parties.seller.ID); err != nil {
			t.Fatalf("Sellers.Delete(whose only sale rolled back) error = %v", err)
		}
		retried := newSale(parties, "key-1")
		retried.Seller = nil
		if again := mustCreateSale(t, repositories.Sales, retried); again.Number < 1 {
			t.Fatalf("number after a rollback = %d", again.Number)
		}
	})

	t.Run("concurrent sales get distinct numbers", func(t *testing.T) {
		repositories := newRepositories(t)
		parties := newSaleParties(t, repositories)
		const workers = 20
		var wg sync.WaitGroup
		var mu sync.Mutex
		var numbers []int64

		for worker := range workers {
			wg.Go(func() {
				sale, err := repositories.Sales.Create(ctx, newSale(parties, fmt.Sprintf("key-%d", worker)))
				if err != nil {
					t.Errorf("Create() error = %v", err)
					return
				}
				mu.Lock()
				numbers = append(numbers, sale.Number)
				mu.Unlock()
			})
		}
		wg.Wait()

		slices.Sort(numbers)
		for i, number := range numbers {
			if number != int64(i+1) {
				t.Fatalf("numbers = %v, want 1 to %d without repeats", numbers, workers)
			}
		}
	})

	t.Run("concurrent sales with the same key keep only one", func(t *testing.T) {
		repositories := newRepositories(t)
		parties := newSaleParties(t, repositories)
		var wg sync.WaitGroup
		var mu sync.Mutex
		succeeded := 0

		for range 20 {
			wg.Go(func() {
				_, err := repositories.Sales.Create(ctx, newSale(parties, "key-1"))
				if err != nil && !errors.Is(err, domain.ErrIdempotencyKeyConflict) {
					t.Errorf("Create() error = %v, want nil or ErrIdempotencyKeyConflict", err)
				}
				if err == nil {
					mu.Lock()
					succeeded++
					mu.Unlock()
				}
			})
		}
		wg.Wait()

		if succeeded != 1 {
			t.Fatalf("%d concurrent creates with one key succeeded, want 1", succeeded)
		}
	})

	t.Run("callers cannot mutate stored sales", func(t *testing.T) {
		repositories := newRepositories(t)
		sale := newSale(newSaleParties(t, repositories), "key-1")
		want := describeSale(sale)

		created := mustCreateSale(t, repositories.Sales, sale)
		sale.Lines[0].Quantity, *sale.Payments[0].Reference, sale.Seller.Name = 9, "changed", "changed"
		sale.Lines[0].Discount.Value, sale.CartDiscount.Value = 1, 1
		created.Lines[1].LineTotal, created.Payments[1].Amount, created.Seller.Code = 1, 1, "XX"
		*created.Payments[0].Reference = "changed"
		created.Lines[1].Discount.Value, created.CartDiscount.Type = 1, domain.DiscountAmount

		got, _ := repositories.Sales.GetByIdempotencyKey(ctx, "u-contract", "key-1")
		got.ID, got.Number = "", 0
		if describeSale(got) != want {
			t.Fatalf("stored sale = %s\nwant %s", describeSale(got), want)
		}
	})

	t.Run("get returns a sale and unknown ids are not found", func(t *testing.T) {
		repositories := newRepositories(t)
		created := mustCreateSale(t, repositories.Sales, newSale(newSaleParties(t, repositories), "key-1"))

		got, err := repositories.Sales.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if describeSale(got) != describeSale(created) {
			t.Fatalf("Get() = %s\nwant %s", describeSale(got), describeSale(created))
		}
		if _, err := repositories.Sales.Get(ctx, "00000000-0000-7000-8000-000000000000"); !errors.Is(err, domain.ErrSaleNotFound) {
			t.Fatalf("Get(unknown) error = %v, want ErrSaleNotFound", err)
		}
	})

	t.Run("void stores the void and nothing else", func(t *testing.T) {
		repositories := newRepositories(t)
		created := mustCreateSale(t, repositories.Sales, newSale(newSaleParties(t, repositories), "key-1"))
		voided, err := created.Void("Error de carga", domain.Actor{ID: "u-manager"}, contractTime.Add(time.Hour))
		if err != nil {
			t.Fatalf("Sale.Void() error = %v", err)
		}
		tampered := voided
		tampered.Total, tampered.Lines = 1, nil

		stored, err := repositories.Sales.Void(ctx, tampered)
		if err != nil {
			t.Fatalf("Void() error = %v", err)
		}

		got, _ := repositories.Sales.Get(ctx, created.ID)
		if describeSale(stored) != describeSale(voided) || describeSale(got) != describeSale(voided) {
			t.Fatalf("Void() = %s\nGet() = %s\nwant %s", describeSale(stored), describeSale(got), describeSale(voided))
		}
		*stored.VoidReason, *stored.VoidedBy = "changed", "changed"
		*voided.VoidedAt = contractTime
		again, _ := repositories.Sales.Get(ctx, created.ID)
		if *again.VoidReason != "Error de carga" || *again.VoidedBy != "u-manager" || !again.VoidedAt.Equal(contractTime.Add(time.Hour)) {
			t.Fatalf("stored void = %s", describeSale(again))
		}

		unknown := voided
		unknown.ID = "00000000-0000-7000-8000-000000000000"
		if _, err := repositories.Sales.Void(ctx, unknown); !errors.Is(err, domain.ErrSaleNotFound) {
			t.Fatalf("Void(unknown) error = %v, want ErrSaleNotFound", err)
		}
	})

	t.Run("a failed unit of work keeps the sale completed", func(t *testing.T) {
		repositories := newRepositories(t)
		created := mustCreateSale(t, repositories.Sales, newSale(newSaleParties(t, repositories), "key-1"))
		voided, _ := created.Void("Error de carga", domain.Actor{ID: "u-manager"}, contractTime.Add(time.Hour))
		failure := errors.New("the stock could not be restored")

		err := repositories.Tx.WithinTx(ctx, func(ctx context.Context) error {
			if _, err := repositories.Sales.Void(ctx, voided); err != nil {
				return err
			}
			return failure
		})
		if !errors.Is(err, failure) {
			t.Fatalf("WithinTx() error = %v, want %v", err, failure)
		}

		got, _ := repositories.Sales.Get(ctx, created.ID)
		if describeSale(got) != describeSale(created) {
			t.Fatalf("after the rollback = %s\nwant %s", describeSale(got), describeSale(created))
		}
	})

	t.Run("list filters by period, customer, seller and status, newest first", func(t *testing.T) {
		repositories := newRepositories(t)
		parties := newSaleParties(t, repositories)
		beto := mustCreateCustomer(t, repositories.Customers, newCustomer(t, "Beto", domain.DocumentNone, ""))
		lucia := mustCreateSeller(t, repositories.Sellers, newSeller(t, "Lucía", "LF"))
		create := func(key string, at time.Time, change func(*domain.Sale)) domain.Sale {
			sale := newSale(parties, key)
			sale.CreatedAt = at
			if change != nil {
				change(&sale)
			}
			return mustCreateSale(t, repositories.Sales, sale)
		}
		first := create("key-1", contractTime, nil)
		second := create("key-2", contractTime.Add(time.Hour), func(sale *domain.Sale) {
			sale.Customer.ID, sale.Customer.Name = beto.ID, beto.Name
		})
		third := create("key-3", contractTime.Add(2*time.Hour), func(sale *domain.Sale) {
			sale.Seller = &domain.SaleSeller{ID: lucia.ID, Name: lucia.Name, Code: lucia.Code}
		})
		fourth := create("key-4", contractTime.Add(3*time.Hour), func(sale *domain.Sale) { sale.Seller = nil })
		voided, _ := third.Void("Error", domain.Actor{ID: "u-manager"}, contractTime.Add(4*time.Hour))
		if _, err := repositories.Sales.Void(ctx, voided); err != nil {
			t.Fatalf("Void() error = %v", err)
		}
		all := domain.PageRequest{Page: 1, Limit: 10}

		tests := []struct {
			name   string
			filter domain.SaleFilter
			page   domain.PageRequest
			want   []domain.Sale
			total  int
		}{
			{name: "all", want: []domain.Sale{fourth, third, second, first}},
			{name: "from, included", filter: domain.SaleFilter{From: contractTime.Add(time.Hour)}, want: []domain.Sale{fourth, third, second}},
			{name: "to, excluded", filter: domain.SaleFilter{To: contractTime.Add(2 * time.Hour)}, want: []domain.Sale{second, first}},
			{name: "a period", filter: domain.SaleFilter{From: contractTime.Add(time.Minute), To: contractTime.Add(3 * time.Hour)}, want: []domain.Sale{third, second}},
			{name: "customer", filter: domain.SaleFilter{CustomerID: beto.ID}, want: []domain.Sale{second}},
			{name: "seller", filter: domain.SaleFilter{SellerID: parties.seller.ID}, want: []domain.Sale{second, first}},
			{name: "status", filter: domain.SaleFilter{Status: domain.SaleVoided}, want: []domain.Sale{third}},
			{name: "seller and status", filter: domain.SaleFilter{SellerID: lucia.ID, Status: domain.SaleCompleted}, want: []domain.Sale{}},
			{name: "second page", page: domain.PageRequest{Page: 2, Limit: 3}, want: []domain.Sale{first}, total: 4},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				page := test.page
				if page.Page == 0 {
					page = all
				}
				got, err := repositories.Sales.List(ctx, gateway.SaleFilter{SaleFilter: test.filter, Page: page})
				if err != nil {
					t.Fatalf("List() error = %v", err)
				}
				total := test.total
				if total == 0 {
					total = len(test.want)
				}
				if fmt.Sprint(saleNumbers(got.Items)) != fmt.Sprint(saleNumbers(test.want)) || got.Total != total || got.Items == nil ||
					got.Page != page.Page || got.Limit != page.Limit {
					t.Fatalf("List() = %v (total %d), want %v (total %d)", saleNumbers(got.Items), got.Total, saleNumbers(test.want), total)
				}
			})
		}

		listed, _ := repositories.Sales.List(ctx, gateway.SaleFilter{SaleFilter: domain.SaleFilter{Status: domain.SaleVoided}, Page: all})
		if describeSale(listed.Items[0]) != describeSale(voided) {
			t.Fatalf("listed = %s\nwant %s", describeSale(listed.Items[0]), describeSale(voided))
		}
	})
}
