package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	"github.com/savvinovan/kyc-service/config"
	appkyc "github.com/savvinovan/kyc-service/internal/application/kyc"
	"github.com/savvinovan/kyc-service/internal/application/command"
	"github.com/savvinovan/kyc-service/internal/application/query"
	"github.com/savvinovan/kyc-service/internal/infrastructure/eventstore"
	httpinterface "github.com/savvinovan/kyc-service/internal/interfaces/http"
	"github.com/savvinovan/kyc-service/internal/interfaces/http/handler"
)

func main() {
	fx.New(
		fx.Provide(
			config.Load,
			newLogger,

			// Infrastructure
			newEventStore,

			// Application — command handlers
			appkyc.NewSubmitKYCHandler,
			appkyc.NewApproveKYCHandler,
			appkyc.NewRejectKYCHandler,

			// Application — query handlers
			appkyc.NewGetKYCStatusHandler,

			// Buses
			newCommandBus,
			newQueryBus,

			// HTTP
			handler.NewHealthHandler,
			handler.NewKYCHandler,
			httpinterface.NewRouter,
			newHTTPServer,
		),
		fx.Invoke(startHTTPServer),
	).Run()
}

// newCommandBus wires all command handlers into the in-memory command bus.
func newCommandBus(
	submit *appkyc.SubmitKYCHandler,
	approve *appkyc.ApproveKYCHandler,
	reject *appkyc.RejectKYCHandler,
) command.Bus {
	bus := command.NewInMemoryBus()
	command.MustRegister(bus, submit)
	command.MustRegister(bus, approve)
	command.MustRegister(bus, reject)
	return bus
}

// newQueryBus wires all query handlers into the in-memory query bus.
func newQueryBus(
	getStatus *appkyc.GetKYCStatusHandler,
) query.Bus {
	bus := query.NewInMemoryBus()
	query.MustRegister(bus, getStatus)
	return bus
}

func newEventStore() eventstore.EventStore {
	return eventstore.NewInMemory()
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
