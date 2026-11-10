package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"market/backend/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := app.New()
	if err != nil {
		slog.Error("backend startup failed", "error", err)
		os.Exit(1)
	}

	if err := application.Run(ctx); err != nil {
		slog.Error("backend stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
