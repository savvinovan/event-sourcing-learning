//go:build integration

package helpers

import (
	"context"
	"database/sql"
	"embed"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredpanda "github.com/testcontainers/testcontainers-go/modules/redpanda"
	"github.com/testcontainers/testcontainers-go/wait"
)

// StartPostgres starts a PostgreSQL 16 container, runs goose migrations, and returns
// a ready pgxpool.Pool, the DSN string, and a cleanup function.
func StartPostgres(ctx context.Context, t TB, migrations embed.FS) (*pgxpool.Pool, string, func()) {
	t.Helper()

	c, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("wallet_test"),
		tcpostgres.WithUsername("wallet"),
		tcpostgres.WithPassword("wallet"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	dsn, err := c.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("get postgres connection string: %v", err)
	}

	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db for migrations: %v", err)
	}
	goose.SetBaseFS(migrations)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("goose set dialect: %v", err)
	}
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	_ = sqlDB.Close()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create pgxpool: %v", err)
	}

	return pool, dsn, func() {
		pool.Close()
		_ = c.Terminate(ctx)
	}
}

// StartRedpanda starts a Redpanda (Kafka-compatible) container and returns
// the seed broker address (host:port) and a cleanup function.
func StartRedpanda(ctx context.Context, t TB) (string, func()) {
	t.Helper()

	c, err := tcredpanda.Run(ctx, "docker.redpanda.com/redpandadata/redpanda:v23.3.3")
	if err != nil {
		t.Fatalf("start redpanda container: %v", err)
	}

	broker, err := c.KafkaSeedBroker(ctx)
	if err != nil {
		t.Fatalf("get redpanda seed broker: %v", err)
	}

	return broker, func() {
		_ = c.Terminate(ctx)
	}
}
