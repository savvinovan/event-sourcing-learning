package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"go.uber.org/fx"

	appaccount "github.com/savvinovan/wallet-service/internal/application/account"
	"github.com/savvinovan/wallet-service/config"
	"github.com/savvinovan/wallet-service/db"
	"github.com/savvinovan/wallet-service/internal/infrastructure/eventstore"
	"github.com/savvinovan/wallet-service/internal/infrastructure/kafka"
	"github.com/savvinovan/wallet-service/internal/infrastructure/readmodel"
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

func newDBPool(cfg *config.Config, lc fx.Lifecycle) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DB.DSN)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			pool.Close()
			return nil
		},
	})
	return pool, nil
}

func newEventStore(pool *pgxpool.Pool) eventstore.EventStore {
	return eventstore.NewPostgresEventStore(pool, eventstore.NewAccountRegistry())
}

func newReadModelRepo(pool *pgxpool.Pool) appaccount.AccountReadRepository {
	return readmodel.NewPostgresReadModelRepository(pool)
}

func runMigrations(cfg *config.Config, log *slog.Logger) error {
	sqlDB, err := sql.Open("pgx", cfg.DB.DSN)
	if err != nil {
		return fmt.Errorf("open db for migrations: %w", err)
	}
	defer sqlDB.Close()

	goose.SetBaseFS(db.Migrations)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	log.Info("migrations applied")
	return nil
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

func startKYCConsumer(lc fx.Lifecycle, consumer *kafka.KYCEventConsumer, log *slog.Logger) {
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go consumer.Run(ctx)
			return nil
		},
		OnStop: func(_ context.Context) error {
			log.Info("stopping KYC event consumer")
			cancel()
			return consumer.Close()
		},
	})
}
