// Command migrate applies the SQL migrations under migrations/ to the
// database named by DATABASE_URL, using Goose. The first argument is the
// Goose command to run ("up", "status", "down", …); it defaults to "up".
package main

import (
	"context"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		slog.Error("migrate: DATABASE_URL is not set")
		os.Exit(1)
	}

	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		dir = "../../migrations"
	}

	command := "up"
	args := os.Args[1:]
	if len(args) > 0 {
		command, args = args[0], args[1:]
	}

	db, err := goose.OpenDBWithDriver("postgres", databaseURL)
	if err != nil {
		slog.Error("migrate: open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := goose.RunContext(context.Background(), command, db, dir, args...); err != nil {
		slog.Error("migrate: run", "command", command, "error", err)
		os.Exit(1)
	}
}
