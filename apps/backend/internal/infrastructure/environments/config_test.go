package environments

import (
	"strings"
	"testing"
)

func clearEnvironment(t *testing.T) {
	t.Helper()
	for _, name := range []string{"PORT", "STORAGE_DRIVER", "DATABASE_URL", "CORS_ALLOWED_ORIGIN", "SEED_DEMO_DATA", "STOCK_ALLOW_NEGATIVE"} {
		t.Setenv(name, "")
	}
}

func TestLoadServerDefaults(t *testing.T) {
	clearEnvironment(t)

	cfg, err := LoadServer()
	if err != nil {
		t.Fatalf("LoadServer() error = %v", err)
	}

	want := Server{Port: "8080", StorageDriver: StorageMemory, SeedDemoData: true}
	if cfg != want {
		t.Fatalf("LoadServer() = %+v, want %+v", cfg, want)
	}
}

func TestLoadServerReadsTheEnvironment(t *testing.T) {
	clearEnvironment(t)
	t.Setenv("PORT", "9090")
	t.Setenv("STORAGE_DRIVER", "memory")
	t.Setenv("CORS_ALLOWED_ORIGIN", " http://localhost:5173/ ")
	t.Setenv("SEED_DEMO_DATA", "false")
	t.Setenv("STOCK_ALLOW_NEGATIVE", "true")

	cfg, err := LoadServer()
	if err != nil {
		t.Fatalf("LoadServer() error = %v", err)
	}

	want := Server{Port: "9090", AllowedOrigin: "http://localhost:5173", StorageDriver: StorageMemory, SeedDemoData: false, StockAllowNegative: true}
	if cfg != want {
		t.Fatalf("LoadServer() = %+v, want %+v", cfg, want)
	}
}

func TestLoadServerAcceptsPostgresWithADatabaseURL(t *testing.T) {
	clearEnvironment(t)
	t.Setenv("STORAGE_DRIVER", "postgres")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/market")

	cfg, err := LoadServer()
	if err != nil {
		t.Fatalf("LoadServer() error = %v", err)
	}

	want := Server{Port: "8080", StorageDriver: StoragePostgres, DatabaseURL: "postgres://user:pass@localhost:5432/market", SeedDemoData: true}
	if cfg != want {
		t.Fatalf("LoadServer() = %+v, want %+v", cfg, want)
	}
}

func TestLoadServerRejectsPostgresWithoutADatabaseURL(t *testing.T) {
	clearEnvironment(t)
	t.Setenv("STORAGE_DRIVER", "postgres")

	if _, err := LoadServer(); err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("LoadServer() error = %v, want a DATABASE_URL error", err)
	}
}

func TestLoadServerReportsEveryProblem(t *testing.T) {
	clearEnvironment(t)
	t.Setenv("PORT", "70000")
	t.Setenv("STORAGE_DRIVER", "sqlite")
	t.Setenv("SEED_DEMO_DATA", "maybe")
	t.Setenv("STOCK_ALLOW_NEGATIVE", "sometimes")

	_, err := LoadServer()
	if err == nil {
		t.Fatal("LoadServer() error = nil, want an error")
	}
	for _, name := range []string{"PORT", "STORAGE_DRIVER", "SEED_DEMO_DATA", "STOCK_ALLOW_NEGATIVE"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q does not mention %s", err, name)
		}
	}
}

func TestLoadServerRejectsANonNumericPort(t *testing.T) {
	clearEnvironment(t)
	t.Setenv("PORT", "http")

	if _, err := LoadServer(); err == nil || !strings.Contains(err.Error(), "PORT") {
		t.Fatalf("LoadServer() error = %v, want a PORT error", err)
	}
}
