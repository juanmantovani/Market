package environments

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

type StorageDriver string

const (
	// StorageMemory keeps everything in the process: data is lost on
	// restart.
	StorageMemory StorageDriver = "memory"
	// StoragePostgres persists through PostgreSQL. DatabaseURL names the
	// database to connect to.
	StoragePostgres StorageDriver = "postgres"
)

type Server struct {
	Port string
	// AllowedOrigin enables CORS for a frontend served from another origin.
	// Empty when frontend and API share an origin (Vite proxy, reverse
	// proxy or the backend serving the build).
	AllowedOrigin string
	StorageDriver StorageDriver
	// DatabaseURL is the PostgreSQL connection string. Required when
	// StorageDriver is StoragePostgres, ignored otherwise.
	DatabaseURL string
	// SeedDemoData loads sample records on startup. It defaults to true
	// because in-memory storage starts empty on every restart.
	SeedDemoData bool
	// StockAllowNegative lets a sale take more stock than there is. It
	// defaults to false: the sale fails with insufficient_stock.
	StockAllowNegative bool
}

// LoadServer reads the server configuration from the environment and
// reports every invalid variable at once, by name only, so the error is safe
// to log.
func LoadServer() (Server, error) {
	var problems []string

	port := envOrDefault("PORT", "8080")
	if number, err := strconv.Atoi(port); err != nil || number < 1 || number > 65535 {
		problems = append(problems, "PORT must be a number between 1 and 65535")
	}

	driver := StorageDriver(envOrDefault("STORAGE_DRIVER", string(StorageMemory)))
	if driver != StorageMemory && driver != StoragePostgres {
		problems = append(problems, `STORAGE_DRIVER must be "memory" or "postgres"`)
	}

	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if driver == StoragePostgres && databaseURL == "" {
		problems = append(problems, "DATABASE_URL is required when STORAGE_DRIVER is \"postgres\"")
	}

	seed, err := strconv.ParseBool(envOrDefault("SEED_DEMO_DATA", "true"))
	if err != nil {
		problems = append(problems, "SEED_DEMO_DATA must be true or false")
	}

	allowNegative, err := strconv.ParseBool(envOrDefault("STOCK_ALLOW_NEGATIVE", "false"))
	if err != nil {
		problems = append(problems, "STOCK_ALLOW_NEGATIVE must be true or false")
	}

	if len(problems) > 0 {
		return Server{}, errors.New(strings.Join(problems, "; "))
	}

	return Server{
		Port:               port,
		AllowedOrigin:      strings.TrimRight(strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGIN")), "/"),
		StorageDriver:      driver,
		DatabaseURL:        databaseURL,
		SeedDemoData:       seed,
		StockAllowNegative: allowNegative,
	}, nil
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}

	return fallback
}
