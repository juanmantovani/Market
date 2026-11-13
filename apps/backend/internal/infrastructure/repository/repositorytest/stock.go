package repositorytest

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/business/gateway"
)

func newMovement(itemID string, movementType domain.MovementType, quantity, balance domain.Quantity, at time.Time) domain.StockMovement {
	return domain.StockMovement{ItemID: itemID, Type: movementType, Quantity: quantity, BalanceAfter: balance, ActorID: "u-contract", CreatedAt: at}
}

func mustRecord(t *testing.T, repository gateway.StockRepository, movement domain.StockMovement) domain.StockMovement {
	t.Helper()
	recorded, err := repository.Record(context.Background(), movement)
	if err != nil {
		t.Fatalf("Record(%s %s) error = %v", movement.Type, movement.ItemID, err)
	}
	return recorded
}

func describeMovement(movement domain.StockMovement) string {
	deref := func(value *string) string {
		if value == nil {
			return "<nil>"
		}
		return *value
	}
	cost := "<nil>"
	if movement.UnitCost != nil {
		cost = fmt.Sprint(*movement.UnitCost)
	}
	return fmt.Sprintf("%s %s %s %d → %d reason=%s reference=%s supplier=%s cost=%s by %s at %s",
		movement.ID, movement.ItemID, movement.Type, movement.Quantity, movement.BalanceAfter,
		deref(movement.Reason), deref(movement.Reference), deref(movement.SupplierID), cost,
		movement.ActorID, movement.CreatedAt.UTC().Format(time.RFC3339Nano))
}

func movementIDs(movements []domain.StockMovement) []string {
	result := make([]string, len(movements))
	for i, movement := range movements {
		result[i] = movement.ID
	}
	return result
}

func stockNames(levels []domain.ItemStock) []string {
	result := make([]string, len(levels))
	for i, level := range levels {
		result[i] = fmt.Sprintf("%s=%d", level.Item.Name, level.Level.Quantity)
	}
	return result
}

// StockRepository runs the gateway.StockRepository contract, including how
// it shares the unit of work of the adapter's TxManager and how movements
// keep items and suppliers from being deleted.
func StockRepository(t *testing.T, newRepositories NewRepositories) {
	ctx := context.Background()

	setup := func(t *testing.T) (Repositories, domain.Item) {
		t.Helper()
		repositories := newRepositories(t)
		almacen := mustCreate(t, repositories.Categories, newCategory(t, "Almacén", nil))
		yerba := mustCreateItem(t, repositories.Items, newItem(t, almacen.ID, "YER-1", "Yerba"))
		return repositories, yerba
	}

	t.Run("an item without movements has a zero level", func(t *testing.T) {
		repositories, yerba := setup(t)

		level, err := repositories.Stock.GetLevel(ctx, yerba.ID)
		if err != nil {
			t.Fatalf("GetLevel() error = %v", err)
		}
		if level != (domain.StockLevel{ItemID: yerba.ID}) || level.HasMovements() {
			t.Fatalf("GetLevel() = %+v, want a zero level", level)
		}
	})

	t.Run("record stores every field and leaves the level at the balance", func(t *testing.T) {
		repositories, yerba := setup(t)
		centro := mustCreateSupplier(t, repositories.Suppliers, newSupplier(t, "Distribuidora del Centro", nil))

		initial := mustRecord(t, repositories.Stock, newMovement(yerba.ID, domain.MovementInitial, 10_000, 10_000, contractTime))
		entry := newMovement(yerba.ID, domain.MovementEntry, 5_000, 15_000, contractTime.Add(time.Hour))
		entry.SupplierID = text(centro.ID)
		entry.Reference = text("Remito 12")
		cost := domain.Money(310_000)
		entry.UnitCost = &cost
		received := mustRecord(t, repositories.Stock, entry)
		adjustment := newMovement(yerba.ID, domain.MovementAdjustment, -1_000, 14_000, contractTime.Add(2*time.Hour))
		adjustment.Reason = text("Rotura")
		adjusted := mustRecord(t, repositories.Stock, adjustment)

		if initial.ID == "" || received.ID == "" || initial.ID == received.ID {
			t.Fatalf("Record() IDs = %q, %q", initial.ID, received.ID)
		}
		level, err := repositories.Stock.GetLevel(ctx, yerba.ID)
		if err != nil {
			t.Fatalf("GetLevel() error = %v", err)
		}
		if level.ItemID != yerba.ID || level.Quantity != 14_000 || !level.UpdatedAt.Equal(contractTime.Add(2*time.Hour)) {
			t.Fatalf("GetLevel() = %+v", level)
		}

		page, err := repositories.Stock.ListMovements(ctx, gateway.StockMovementFilter{ItemID: yerba.ID, Page: domain.PageRequest{Page: 1, Limit: 10}})
		if err != nil {
			t.Fatalf("ListMovements() error = %v", err)
		}
		want := []domain.StockMovement{adjusted, received, initial}
		if len(page.Items) != len(want) {
			t.Fatalf("ListMovements() = %v", movementIDs(page.Items))
		}
		for i := range want {
			if got := describeMovement(page.Items[i]); got != describeMovement(want[i]) {
				t.Fatalf("movement %d = %s, want %s", i, got, describeMovement(want[i]))
			}
		}
	})

	t.Run("record needs the item and the supplier to exist", func(t *testing.T) {
		repositories, yerba := setup(t)
		unknown := "00000000-0000-7000-8000-000000000000"

		if _, err := repositories.Stock.Record(ctx, newMovement(unknown, domain.MovementInitial, 1_000, 1_000, contractTime)); !errors.Is(err, domain.ErrStockItemNotFound) {
			t.Fatalf("Record(unknown item) error = %v, want ErrStockItemNotFound", err)
		}
		entry := newMovement(yerba.ID, domain.MovementEntry, 1_000, 1_000, contractTime)
		entry.SupplierID = text(unknown)
		if _, err := repositories.Stock.Record(ctx, entry); !errors.Is(err, domain.ErrStockSupplierNotFound) {
			t.Fatalf("Record(unknown supplier) error = %v, want ErrStockSupplierNotFound", err)
		}

		level, _ := repositories.Stock.GetLevel(ctx, yerba.ID)
		movements, _ := repositories.Stock.ListMovements(ctx, gateway.StockMovementFilter{Page: domain.PageRequest{Page: 1, Limit: 10}})
		if level.HasMovements() || movements.Total != 0 {
			t.Fatalf("a rejected movement left level %+v and %d movements", level, movements.Total)
		}
	})

	t.Run("items and suppliers with movements cannot be deleted", func(t *testing.T) {
		repositories, yerba := setup(t)
		centro := mustCreateSupplier(t, repositories.Suppliers, newSupplier(t, "Distribuidora del Centro", nil))
		entry := newMovement(yerba.ID, domain.MovementEntry, 1_000, 1_000, contractTime)
		entry.SupplierID = text(centro.ID)
		mustRecord(t, repositories.Stock, entry)

		if err := repositories.Items.Delete(ctx, yerba.ID); !errors.Is(err, domain.ErrItemInUse) {
			t.Fatalf("Items.Delete(with movements) error = %v, want ErrItemInUse", err)
		}
		if err := repositories.Suppliers.Delete(ctx, centro.ID); !errors.Is(err, domain.ErrSupplierInUse) {
			t.Fatalf("Suppliers.Delete(with entries) error = %v, want ErrSupplierInUse", err)
		}

		arroz := mustCreateItem(t, repositories.Items, newItem(t, yerba.CategoryID, "ARR-1", "Arroz"))
		if err := repositories.Items.Delete(ctx, arroz.ID); err != nil {
			t.Fatalf("Items.Delete(without movements) error = %v", err)
		}
	})

	t.Run("list levels lists every item, searches and filters by category and low stock", func(t *testing.T) {
		repositories := newRepositories(t)
		almacen := mustCreate(t, repositories.Categories, newCategory(t, "Almacén", nil))
		bebidas := mustCreate(t, repositories.Categories, newCategory(t, "Bebidas", nil))
		create := func(categoryID, code, name string, minStock, stock domain.Quantity) domain.Item {
			item := newItem(t, categoryID, code, name)
			item.MinStock = minStock
			created := mustCreateItem(t, repositories.Items, item)
			if stock != 0 {
				mustRecord(t, repositories.Stock, newMovement(created.ID, domain.MovementInitial, stock, stock, contractTime))
			}
			return created
		}
		yerba := create(almacen.ID, "YER-1", "Yerba mate", 6_000, 4_000)
		arroz := create(almacen.ID, "ARR-1", "Arroz", 10_000, 25_000)
		create(almacen.ID, "FID-1", "fideos", 5_000, 5_000)
		coca := create(bebidas.ID, "COC-1", "Coca-Cola", 0, 0)
		page := domain.PageRequest{Page: 1, Limit: 10}

		tests := []struct {
			name   string
			filter gateway.StockLevelFilter
			want   []string
		}{
			{name: "all", filter: gateway.StockLevelFilter{Page: page}, want: []string{"Arroz=25000", "Coca-Cola=0", "fideos=5000", "Yerba mate=4000"}},
			{name: "low stock, the minimum included", filter: gateway.StockLevelFilter{LowStock: true, Page: page}, want: []string{"fideos=5000", "Yerba mate=4000"}},
			{name: "by category", filter: gateway.StockLevelFilter{CategoryID: bebidas.ID, Page: page}, want: []string{"Coca-Cola=0"}},
			{name: "by name without accents", filter: gateway.StockLevelFilter{Search: "YERBA", Page: page}, want: []string{"Yerba mate=4000"}},
			{name: "by code", filter: gateway.StockLevelFilter{Search: "arr-1", Page: page}, want: []string{"Arroz=25000"}},
			{name: "low stock within a category", filter: gateway.StockLevelFilter{CategoryID: bebidas.ID, LowStock: true, Page: page}, want: []string{}},
			{name: "by item IDs", filter: gateway.StockLevelFilter{ItemIDs: []string{yerba.ID, coca.ID}, Page: page}, want: []string{"Coca-Cola=0", "Yerba mate=4000"}},
			{name: "by item IDs, ignoring unknown ones", filter: gateway.StockLevelFilter{ItemIDs: []string{arroz.ID, "unknown"}, Page: page}, want: []string{"Arroz=25000"}},
			{name: "only unknown item IDs", filter: gateway.StockLevelFilter{ItemIDs: []string{"unknown"}, Page: page}, want: []string{}},
			{name: "by item IDs and low stock", filter: gateway.StockLevelFilter{ItemIDs: []string{yerba.ID, arroz.ID}, LowStock: true, Page: page}, want: []string{"Yerba mate=4000"}},
			{name: "by item IDs and category", filter: gateway.StockLevelFilter{ItemIDs: []string{yerba.ID, coca.ID}, CategoryID: bebidas.ID, Page: page}, want: []string{"Coca-Cola=0"}},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				got, err := repositories.Stock.ListLevels(ctx, test.filter)
				if err != nil {
					t.Fatalf("ListLevels() error = %v", err)
				}
				if fmt.Sprint(stockNames(got.Items)) != fmt.Sprint(test.want) || got.Total != len(test.want) || got.Items == nil {
					t.Fatalf("ListLevels() = %v (total %d), want %v", stockNames(got.Items), got.Total, test.want)
				}
			})
		}

		for name, filter := range map[string]gateway.ItemFilter{
			"low stock":                  {LowStock: true, Page: page},
			"low stock in its category":  {LowStock: true, CategoryID: almacen.ID, Page: page},
			"low stock in another one":   {LowStock: true, CategoryID: bebidas.ID, Page: page},
			"low stock matching a name":  {LowStock: true, Search: "yerba", Page: page},
			"low stock not matching one": {LowStock: true, Search: "arroz", Page: page},
		} {
			items, err := repositories.Items.List(ctx, filter)
			if err != nil {
				t.Fatalf("Items.List(%s) error = %v", name, err)
			}
			want := map[string][]string{
				"low stock": {"fideos", "Yerba mate"}, "low stock in its category": {"fideos", "Yerba mate"},
				"low stock in another one": {}, "low stock matching a name": {"Yerba mate"}, "low stock not matching one": {},
			}[name]
			if fmt.Sprint(itemNames(items.Items)) != fmt.Sprint(want) || items.Total != len(want) {
				t.Fatalf("Items.List(%s) = %v (total %d), want %v", name, itemNames(items.Items), items.Total, want)
			}
		}

		got, _ := repositories.Stock.ListLevels(ctx, gateway.StockLevelFilter{CategoryID: bebidas.ID, Page: page})
		if level := got.Items[0]; level.Item.ID != coca.ID || level.Level != (domain.StockLevel{ItemID: coca.ID}) || level.Item.Code != "COC-1" {
			t.Fatalf("level of an item without movements = %+v", level)
		}
	})

	t.Run("list levels pages through the results", func(t *testing.T) {
		repositories, yerba := setup(t)
		for _, name := range []string{"E", "A", "D", "C"} {
			mustCreateItem(t, repositories.Items, newItem(t, yerba.CategoryID, "CODE-"+name, name))
		}

		second, err := repositories.Stock.ListLevels(ctx, gateway.StockLevelFilter{Page: domain.PageRequest{Page: 2, Limit: 2}})
		if err != nil {
			t.Fatalf("ListLevels() error = %v", err)
		}
		if fmt.Sprint(stockNames(second.Items)) != "[D=0 E=0]" || second.Total != 5 || second.Page != 2 || second.Limit != 2 {
			t.Fatalf("page 2 = %v %+v", stockNames(second.Items), second)
		}
	})

	t.Run("list movements filters by item and type, newest first, and pages", func(t *testing.T) {
		repositories, yerba := setup(t)
		arroz := mustCreateItem(t, repositories.Items, newItem(t, yerba.CategoryID, "ARR-1", "Arroz"))
		yerbaInitial := mustRecord(t, repositories.Stock, newMovement(yerba.ID, domain.MovementInitial, 4_000, 4_000, contractTime))
		arrozInitial := mustRecord(t, repositories.Stock, newMovement(arroz.ID, domain.MovementInitial, 9_000, 9_000, contractTime))
		yerbaEntry := mustRecord(t, repositories.Stock, newMovement(yerba.ID, domain.MovementEntry, 6_000, 10_000, contractTime.Add(time.Minute)))
		adjustment := newMovement(yerba.ID, domain.MovementAdjustment, -1_000, 9_000, contractTime.Add(2*time.Minute))
		adjustment.Reason = text("Rotura")
		yerbaAdjustment := mustRecord(t, repositories.Stock, adjustment)

		tests := []struct {
			name   string
			filter gateway.StockMovementFilter
			want   []domain.StockMovement
			total  int
		}{
			{name: "all", filter: gateway.StockMovementFilter{}, want: []domain.StockMovement{yerbaAdjustment, yerbaEntry, arrozInitial, yerbaInitial}},
			{name: "by item", filter: gateway.StockMovementFilter{ItemID: yerba.ID}, want: []domain.StockMovement{yerbaAdjustment, yerbaEntry, yerbaInitial}},
			{name: "by type", filter: gateway.StockMovementFilter{Type: domain.MovementInitial}, want: []domain.StockMovement{arrozInitial, yerbaInitial}},
			{name: "by item and type", filter: gateway.StockMovementFilter{ItemID: arroz.ID, Type: domain.MovementEntry}, want: []domain.StockMovement{}},
			{name: "second page", filter: gateway.StockMovementFilter{ItemID: yerba.ID, Page: domain.PageRequest{Page: 2, Limit: 1}}, want: []domain.StockMovement{yerbaEntry}, total: 3},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				if test.filter.Page.Page == 0 {
					test.filter.Page = domain.PageRequest{Page: 1, Limit: 10}
				}
				got, err := repositories.Stock.ListMovements(ctx, test.filter)
				if err != nil {
					t.Fatalf("ListMovements() error = %v", err)
				}
				total := test.total
				if total == 0 {
					total = len(test.want)
				}
				if fmt.Sprint(movementIDs(got.Items)) != fmt.Sprint(movementIDs(test.want)) || got.Total != total || got.Items == nil {
					t.Fatalf("ListMovements() = %v (total %d), want %v", movementIDs(got.Items), got.Total, movementIDs(test.want))
				}
			})
		}
	})

	t.Run("a failed unit of work records nothing", func(t *testing.T) {
		repositories, yerba := setup(t)
		mustRecord(t, repositories.Stock, newMovement(yerba.ID, domain.MovementInitial, 4_000, 4_000, contractTime))
		arroz := mustCreateItem(t, repositories.Items, newItem(t, yerba.CategoryID, "ARR-1", "Arroz"))
		failure := errors.New("the sale failed")

		err := repositories.Tx.WithinTx(ctx, func(ctx context.Context) error {
			for _, movement := range []domain.StockMovement{
				newMovement(yerba.ID, domain.MovementEntry, 1_000, 5_000, contractTime),
				newMovement(arroz.ID, domain.MovementInitial, 2_000, 2_000, contractTime),
			} {
				if _, err := repositories.Stock.Record(ctx, movement); err != nil {
					return err
				}
			}
			return failure
		})
		if !errors.Is(err, failure) {
			t.Fatalf("WithinTx() error = %v, want %v", err, failure)
		}

		level, _ := repositories.Stock.GetLevel(ctx, yerba.ID)
		movements, _ := repositories.Stock.ListMovements(ctx, gateway.StockMovementFilter{Page: domain.PageRequest{Page: 1, Limit: 10}})
		if level.Quantity != 4_000 || movements.Total != 1 || movements.Items[0].Type != domain.MovementInitial {
			t.Fatalf("after the rollback: level %+v, movements %v", level, movementIDs(movements.Items))
		}
		if err := repositories.Items.Delete(ctx, arroz.ID); err != nil {
			t.Fatalf("Items.Delete(item whose only movement rolled back) error = %v", err)
		}
	})

	t.Run("concurrent units of work never lose a movement", func(t *testing.T) {
		repositories, yerba := setup(t)
		const workers = 20
		var wg sync.WaitGroup
		errs := make(chan error, workers)

		for range workers {
			wg.Go(func() {
				errs <- repositories.Tx.WithinTx(ctx, func(ctx context.Context) error {
					level, err := repositories.Stock.GetLevel(ctx, yerba.ID)
					if err != nil {
						return err
					}
					_, err = repositories.Stock.Record(ctx, newMovement(yerba.ID, domain.MovementEntry, 1_000, level.Quantity+1_000, time.Now().UTC()))
					return err
				})
			})
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			if err != nil {
				t.Fatalf("WithinTx() error = %v", err)
			}
		}

		level, _ := repositories.Stock.GetLevel(ctx, yerba.ID)
		if level.Quantity != workers*1_000 {
			t.Fatalf("level after %d concurrent entries = %d, want %d", workers, level.Quantity, workers*1_000)
		}
		movements, _ := repositories.Stock.ListMovements(ctx, gateway.StockMovementFilter{Page: domain.PageRequest{Page: 1, Limit: 100}})
		balances := make([]domain.Quantity, 0, workers)
		for _, movement := range movements.Items {
			balances = append(balances, movement.BalanceAfter)
		}
		slices.Sort(balances)
		for i, balance := range balances {
			if balance != domain.Quantity(i+1)*1_000 {
				t.Fatalf("balances = %v, want 1000 to %d without repeats", balances, workers*1_000)
			}
		}
	})

	t.Run("callers cannot mutate recorded movements", func(t *testing.T) {
		repositories, yerba := setup(t)
		centro := mustCreateSupplier(t, repositories.Suppliers, newSupplier(t, "Distribuidora del Centro", nil))
		movement := newMovement(yerba.ID, domain.MovementEntry, 1_000, 1_000, contractTime)
		movement.SupplierID = text(centro.ID)
		movement.Reference = text("Remito 12")
		movement.Reason = text("Llegó tarde")
		cost := domain.Money(500)
		movement.UnitCost = &cost

		recorded := mustRecord(t, repositories.Stock, movement)
		*movement.Reference, cost = "changed", 1
		*recorded.Reason, *recorded.SupplierID, *recorded.UnitCost = "changed", "changed", 2

		page, _ := repositories.Stock.ListMovements(ctx, gateway.StockMovementFilter{Page: domain.PageRequest{Page: 1, Limit: 10}})
		listed := page.Items[0]
		if *listed.Reference != "Remito 12" || *listed.Reason != "Llegó tarde" || *listed.SupplierID != centro.ID || *listed.UnitCost != 500 {
			t.Fatalf("stored movement = %s", describeMovement(listed))
		}
	})
}
