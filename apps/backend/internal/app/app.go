package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"market/backend/internal/business/domain"
	"market/backend/internal/infrastructure/delivery/webapp/middleware"
	"market/backend/internal/infrastructure/delivery/webapp/route"
	"market/backend/internal/infrastructure/delivery/webapp/server"
	"market/backend/internal/infrastructure/environments"
)

// devActor acts on every request until an authentication provider is
// chosen.
var devActor = domain.Actor{ID: "dev", Name: "Usuario de desarrollo", Role: domain.RoleAdmin}

type App struct {
	server *http.Server
}

func New() (*App, error) {
	cfg, err := environments.LoadServer()
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	route.RegisterRoutes(mux)

	handler := middleware.RequestLog(
		middleware.Recover(
			server.WithCORS(cfg.AllowedOrigin,
				middleware.DevActor(devActor, mux),
			),
		),
	)

	return &App{server: server.New(cfg.Port, handler, route.MaxHandlerTimeout)}, nil
}

// Run serves requests until ctx is cancelled, then shuts down gracefully.
func (app *App) Run(ctx context.Context) error {
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- app.server.ListenAndServe()
	}()

	slog.Info("backend listening", "address", app.server.Addr)

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := app.server.Shutdown(shutdownCtx); err != nil {
			return err
		}

		err := <-serverErrors
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
