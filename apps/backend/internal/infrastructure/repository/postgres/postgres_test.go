package postgres

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"market/backend/internal/business/domain"
	"market/backend/internal/infrastructure/repository/repositorytest"
)

var (
	testPoolOnce sync.Once
	testPool     *pgxpool.Pool
	testPoolErr  error
)

// requireTestDatabase skips the calling test when there's no PostgreSQL to
// run the contract suite against, so `go test ./...` stays green without
// Docker. CI always sets TEST_DATABASE_URL, pointing at a migrated database.
func requireTestDatabase(t *testing.T) {
	t.Helper()
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL is not set: skipping the PostgreSQL contract suite")
	}
}

func testPoolConn(t *testing.T) *pgxpool.Pool {
	t.Helper()
	testPoolOnce.Do(func() {
		testPool, testPoolErr = pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))
	})
	if testPoolErr != nil {
		t.Fatalf("connect to TEST_DATABASE_URL: %v", testPoolErr)
	}

	return testPool
}

// newRepositories gives every subtest a freshly truncated database, holding
// nothing but Consumidor Final, exactly like memory.NewStore.
func newRepositories(t *testing.T) repositorytest.Repositories {
	t.Helper()
	pool := testPoolConn(t)
	resetTables(t, pool)

	return repositorytest.Repositories{
		Customers: NewCustomerRepository(pool),
		Tx:        NewTxManager(pool),
	}
}

func resetTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	if _, err := pool.Exec(ctx, "TRUNCATE TABLE customers"); err != nil {
		t.Fatalf("truncate customers: %v", err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO customers (id, name, document_type, tax_condition, is_system, created_at, updated_at)
		VALUES ($1, 'Consumidor Final', 'none', 'final_consumer', true, now(), now())`,
		domain.FinalConsumerID); err != nil {
		t.Fatalf("seed Consumidor Final: %v", err)
	}
}
