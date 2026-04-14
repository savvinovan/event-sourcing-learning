package app

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
	"github.com/savvinovan/kyc-service/internal/infrastructure/eventstore"
	"github.com/savvinovan/kyc-service/internal/infrastructure/messaging"
)

func newConfig() (*config.Config, error) {
	return config.Load()
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

func newEventStore() eventstore.EventStore {
	return eventstore.NewInMemory()
}

func newPublisher(cfg *config.Config) appkyc.EventPublisher {
	return messaging.NewKafkaPublisher(cfg.Kafka.Brokers)
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

// registerPublisherShutdown closes the Kafka writer on app stop.
func registerPublisherShutdown(lc fx.Lifecycle, pub appkyc.EventPublisher, log *slog.Logger) {
	type closer interface{ Close() error }
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if c, ok := pub.(closer); ok {
				log.Info("closing Kafka publisher")
				return c.Close()
			}
			return nil
		},
	})
}
