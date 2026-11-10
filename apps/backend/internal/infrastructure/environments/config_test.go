package environments

import (
	"strings"
	"testing"
)

func TestLoadServerDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("STORAGE_DRIVER", "")
	t.Setenv("CORS_ALLOWED_ORIGIN", "")

	cfg, err := LoadServer()
	if err != nil {
		t.Fatalf("LoadServer() error = %v", err)
	}

	want := Server{Port: "8080", StorageDriver: StorageMemory}
	if cfg != want {
		t.Fatalf("LoadServer() = %+v, want %+v", cfg, want)
	}
}

func TestLoadServerReadsTheEnvironment(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("STORAGE_DRIVER", "memory")
	t.Setenv("CORS_ALLOWED_ORIGIN", " http://localhost:5173/ ")

	cfg, err := LoadServer()
	if err != nil {
		t.Fatalf("LoadServer() error = %v", err)
	}

	want := Server{Port: "9090", AllowedOrigin: "http://localhost:5173", StorageDriver: StorageMemory}
	if cfg != want {
		t.Fatalf("LoadServer() = %+v, want %+v", cfg, want)
	}
}

func TestLoadServerReportsEveryProblem(t *testing.T) {
	t.Setenv("PORT", "70000")
	t.Setenv("STORAGE_DRIVER", "postgres")

	_, err := LoadServer()
	if err == nil {
		t.Fatal("LoadServer() error = nil, want an error")
	}
	for _, name := range []string{"PORT", "STORAGE_DRIVER"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q does not mention %s", err, name)
		}
	}
}

func TestLoadServerRejectsANonNumericPort(t *testing.T) {
	t.Setenv("PORT", "http")
	t.Setenv("STORAGE_DRIVER", "")

	if _, err := LoadServer(); err == nil || !strings.Contains(err.Error(), "PORT") {
		t.Fatalf("LoadServer() error = %v, want a PORT error", err)
	}
}
