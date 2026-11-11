package environments

import (
	"strings"
	"testing"
)

func clearEnvironment(t *testing.T) {
	t.Helper()
	for _, name := range []string{"PORT", "STORAGE_DRIVER", "CORS_ALLOWED_ORIGIN", "SEED_DEMO_DATA"} {
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

	cfg, err := LoadServer()
	if err != nil {
		t.Fatalf("LoadServer() error = %v", err)
	}

	want := Server{Port: "9090", AllowedOrigin: "http://localhost:5173", StorageDriver: StorageMemory, SeedDemoData: false}
	if cfg != want {
		t.Fatalf("LoadServer() = %+v, want %+v", cfg, want)
	}
}

func TestLoadServerReportsEveryProblem(t *testing.T) {
	clearEnvironment(t)
	t.Setenv("PORT", "70000")
	t.Setenv("STORAGE_DRIVER", "postgres")
	t.Setenv("SEED_DEMO_DATA", "maybe")

	_, err := LoadServer()
	if err == nil {
		t.Fatal("LoadServer() error = nil, want an error")
	}
	for _, name := range []string{"PORT", "STORAGE_DRIVER", "SEED_DEMO_DATA"} {
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
