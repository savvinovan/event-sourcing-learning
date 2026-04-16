package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // registers "pgx" driver for database/sql (goose)
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"

	"github.com/savvinovan/wallet-service/config"
	"github.com/savvinovan/wallet-service/db"
	"github.com/savvinovan/wallet-service/internal/infrastructure/eventstore"
	"github.com/savvinovan/wallet-service/internal/infrastructure/projector"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		log.Error("load config", "error", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DB.DSN)
	if err != nil {
		log.Error("connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := runMigrations(cfg.DB.DSN, log); err != nil {
		log.Error("run migrations", "error", err)
		os.Exit(1)
	}

	registry := eventstore.NewAccountRegistry()
	upcasters := eventstore.NewAccountUpcasterRegistry()
	applier := projector.NewAccountProjector()
	runner := projector.NewRunner(
		pool,
		registry,
		upcasters,
		applier,
		projector.AccountProjectorName,
		100,                  // batch size
		500*time.Millisecond, // poll interval when idle
		log,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info("wallet projector starting")
	if err := runner.Run(ctx); err != nil {
		log.Error("projector stopped with error", "error", err)
		os.Exit(1)
	}
	log.Info("wallet projector stopped")
}

func runMigrations(dsn string, log *slog.Logger) error {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	goose.SetBaseFS(db.Migrations)
	goose.SetLogger(goose.NopLogger())

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		return err
	}
	log.Info("migrations applied")
	return nil
}
