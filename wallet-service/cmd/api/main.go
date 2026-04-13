package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/savvinovan/wallet-service/config"
	httpinterface "github.com/savvinovan/wallet-service/internal/interfaces/http"
	"github.com/savvinovan/wallet-service/internal/interfaces/http/handler"

	// contracts are used for cross-service event schemas (Kafka integration — see PLAN-006)
	_ "github.com/savvinovan/event-sourcing-learning/contracts/events"
)

func main() {
	fx.New(
		fx.Provide(
			config.Load,
			newLogger,
			handler.NewHealthHandler,
			httpinterface.NewRouter,
			newHTTPServer,
		),
		fx.Invoke(startHTTPServer),
	).Run()
}

func newLogger(cfg *config.Config) *slog.Logger {
	var level slog.Level
	switch cfg.Log.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}

func newHTTPServer(cfg *config.Config, r *chi.Mux) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler: r,
	}
}

func startHTTPServer(lc fx.Lifecycle, srv *http.Server, log *slog.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("starting HTTP server", "addr", srv.Addr)
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Error("HTTP server error", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("stopping HTTP server")
			return srv.Shutdown(ctx)
		},
	})
}
