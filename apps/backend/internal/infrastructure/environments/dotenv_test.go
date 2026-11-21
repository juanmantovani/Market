package environments

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func unsetEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		t.Setenv(key, "placeholder")
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
}

func writeRepo(t *testing.T, env string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "pnpm-workspace.yaml"), []byte("packages:\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if env != "" {
		if err := os.WriteFile(filepath.Join(root, ".env"), []byte(env), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	backend := filepath.Join(root, "apps", "backend")
	if err := os.MkdirAll(backend, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(backend)
	return root
}

func TestLoadProjectDotEnvReadsTheRepositoryFile(t *testing.T) {
	unsetEnv(t, "PORT", "STORAGE_DRIVER", "CORS_ALLOWED_ORIGIN", "SEED_DEMO_DATA", "STOCK_ALLOW_NEGATIVE", "GS1_USERNAME", "GS1_PASSWORD", "MARKET_DOTENV_PROBE")
	writeRepo(t, strings.Join([]string{
		"# comment",
		"",
		"export PORT=9090",
		"STORAGE_DRIVER=memory",
		"SEED_DEMO_DATA=false",
		"STOCK_ALLOW_NEGATIVE=true",
		"CORS_ALLOWED_ORIGIN=",
		"GS1_USERNAME='member'",
		`GS1_PASSWORD="s e#cret"`,
		"MARKET_DOTENV_PROBE=first",
		"MARKET_DOTENV_PROBE=from-file",
		"",
	}, "\r\n"))

	if err := LoadProjectDotEnv(); err != nil {
		t.Fatalf("LoadProjectDotEnv() error = %v", err)
	}

	cfg, err := LoadServer()
	if err != nil {
		t.Fatalf("LoadServer() error = %v", err)
	}
	want := Server{
		Port:               "9090",
		StorageDriver:      StorageMemory,
		SeedDemoData:       false,
		StockAllowNegative: true,
		GS1Username:        "member",
		GS1Password:        "s e#cret",
	}
	if cfg != want {
		t.Fatalf("LoadServer() = %+v, want %+v", cfg, want)
	}
	if got := os.Getenv("MARKET_DOTENV_PROBE"); got != "from-file" {
		t.Fatalf("MARKET_DOTENV_PROBE = %q, want from-file", got)
	}
}

func TestLoadProjectDotEnvKeepsVariablesAlreadySet(t *testing.T) {
	unsetEnv(t, "MARKET_DOTENV_PROBE")
	t.Setenv("MARKET_DOTENV_PROBE", "from-shell")
	writeRepo(t, "MARKET_DOTENV_PROBE=from-file\n")

	if err := LoadProjectDotEnv(); err != nil {
		t.Fatalf("LoadProjectDotEnv() error = %v", err)
	}
	if got := os.Getenv("MARKET_DOTENV_PROBE"); got != "from-shell" {
		t.Fatalf("MARKET_DOTENV_PROBE = %q, want from-shell", got)
	}
}

func TestLoadProjectDotEnvIgnoresAMissingFile(t *testing.T) {
	unsetEnv(t, "MARKET_DOTENV_PROBE")
	writeRepo(t, "")

	if err := LoadProjectDotEnv(); err != nil {
		t.Fatalf("LoadProjectDotEnv() error = %v", err)
	}
	if _, exists := os.LookupEnv("MARKET_DOTENV_PROBE"); exists {
		t.Fatal("MARKET_DOTENV_PROBE was set")
	}
}

func TestLoadProjectDotEnvIgnoresADirectoryOutsideTheRepository(t *testing.T) {
	unsetEnv(t, "MARKET_DOTENV_PROBE")
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("MARKET_DOTENV_PROBE=from-file\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	if err := LoadProjectDotEnv(); err != nil {
		t.Fatalf("LoadProjectDotEnv() error = %v", err)
	}
	if _, exists := os.LookupEnv("MARKET_DOTENV_PROBE"); exists {
		t.Fatal("MARKET_DOTENV_PROBE was set")
	}
}

func TestLoadProjectDotEnvReportsAnInvalidLineWithoutItsContents(t *testing.T) {
	unsetEnv(t, "PORT")
	writeRepo(t, "PORT=9090\nSECRET=\"super-secret-value\n")

	err := LoadProjectDotEnv()
	if err == nil {
		t.Fatal("LoadProjectDotEnv() error = nil, want an error")
	}
	if !strings.Contains(err.Error(), ".env:2") || strings.Contains(err.Error(), "super-secret-value") {
		t.Fatalf("LoadProjectDotEnv() error = %v", err)
	}
	if os.Getenv("PORT") == "9090" {
		t.Fatal("PORT was set from an invalid file")
	}
}

func TestLoadProjectDotEnvReportsAnUnreadableFile(t *testing.T) {
	root := writeRepo(t, "PORT=9090\n")
	path := filepath.Join(root, ".env")
	if err := os.Chmod(path, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(path, 0o644)
	})

	file, openErr := os.Open(path)
	if openErr == nil {
		file.Close()
		t.Skip("process can read a mode 000 file")
	}

	if err := LoadProjectDotEnv(); err == nil {
		t.Fatal("LoadProjectDotEnv() error = nil, want an error")
	}
}
