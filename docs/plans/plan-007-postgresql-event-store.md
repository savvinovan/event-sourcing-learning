# PLAN-007: PostgreSQL Event Store + Async Projector

| | |
|-|-|
| **Status** | DONE |
| **Date** | 2026-04-14 |
| **Depends on** | [PLAN-004](plan-004-wallet-service.md), [PLAN-005](plan-005-kyc-service.md) |

## Goal

1. Replace in-memory `EventStore` with a PostgreSQL-backed implementation.
2. Introduce an **async Projector** — a separate binary that tails the event store and
   builds read model tables. It runs independently from the API, has no HTTP surface,
   and can be scaled, restarted, or replayed without touching the write path.

## Architecture Overview

```
┌──────────────────────────────────┐     ┌──────────────────────────────────┐
│  cmd/wallet-api  (API process)   │     │  cmd/wallet-projector            │
│                                  │     │  (separate binary, separate pod) │
│  HTTP → CommandHandler           │     │                                  │
│           └── EventStore.Append  │     │  event loop:                     │
│               └── INSERT events  │     │    SELECT events WHERE            │
│                                  │     │      global_seq > checkpoint      │
│  HTTP → QueryHandler             │     │    → AccountProjector.Apply()     │
│           └── ReadModelRepo      │     │    → UPSERT account_read_models  │
│               └── SELECT         │◄────│    → INSERT transaction_history  │
│                  account_read_   │     │    → UPDATE checkpoint           │
│                  models / txns   │     │                                  │
└──────────────────────────────────┘     └──────────────────────────────────┘
                    │                                     │
                    └──────────────┬──────────────────────┘
                                   │
                        ┌──────────▼──────────┐
                        │   wallet_db          │
                        │   (PostgreSQL)       │
                        │                      │
                        │  events              │
                        │  account_read_models │
                        │  transaction_history │
                        │  projector_checkpts  │
                        └──────────────────────┘
```

**12-Factor: one database per microservice.**
`wallet-service` owns `wallet_db`. `kyc-service` owns `kyc_db`.
The projector for each service connects to that service's database only.

**Eventual consistency**: there is a brief lag between event append and read model update.
Query handlers may return slightly stale data. This is intentional and documented.

**Why async and not in-transaction?**
An in-transaction projector ties writes to a single DB master and prevents read replica
offloading. The async projector can consume from a read replica, be deployed separately,
be restarted to replay from any checkpoint, and be extended to fan out to multiple
read stores (Redis, Elasticsearch) without touching the write path.

## Database Schema

### Event Store

```sql
-- 001_create_events.sql
CREATE TABLE events (
    global_seq      BIGSERIAL   NOT NULL,           -- global ordering for the projector
    id              UUID        NOT NULL,
    aggregate_id    TEXT        NOT NULL,
    aggregate_type  TEXT        NOT NULL,
    event_type      TEXT        NOT NULL,
    event_version   INT         NOT NULL,
    payload         JSONB       NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (global_seq),
    UNIQUE (aggregate_id, event_version)
);

CREATE INDEX idx_events_aggregate ON events (aggregate_id, event_version ASC);
CREATE INDEX idx_events_seq       ON events (global_seq ASC);
```

`global_seq BIGSERIAL` gives the projector a monotonically increasing cursor.
`UNIQUE (aggregate_id, event_version)` enforces optimistic concurrency at DB level.

### Read Models (wallet-service)

```sql
-- 002_create_read_models.sql

-- Current snapshot of each account — rebuilt by the Projector
CREATE TABLE account_read_models (
    account_id    TEXT            PRIMARY KEY,
    customer_id   TEXT            NOT NULL,
    status        TEXT            NOT NULL,
    balance       NUMERIC(20, 8)  NOT NULL DEFAULT 0,
    currency      TEXT            NOT NULL,
    version       INT             NOT NULL,
    updated_at    TIMESTAMPTZ     NOT NULL
);

-- Append-only ledger of every deposit and withdrawal
CREATE TABLE transaction_history (
    id            UUID            PRIMARY KEY,
    account_id    TEXT            NOT NULL,
    event_type    TEXT            NOT NULL,  -- 'MoneyDeposited' | 'MoneyWithdrawn'
    amount        NUMERIC(20, 8)  NOT NULL,
    currency      TEXT            NOT NULL,
    occurred_at   TIMESTAMPTZ     NOT NULL
);

CREATE INDEX idx_tx_account ON transaction_history (account_id, occurred_at ASC);
```

`NUMERIC(20, 8)` mirrors `decimal.Decimal` precision — never `FLOAT` or `DOUBLE PRECISION`.

### Projector Checkpoint

```sql
-- 003_create_projector_checkpoint.sql
CREATE TABLE projector_checkpoints (
    projector_name  TEXT        PRIMARY KEY,
    last_global_seq BIGINT      NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- seed row so the projector can always UPDATE (never INSERT on hot path)
INSERT INTO projector_checkpoints (projector_name, last_global_seq)
VALUES ('account_projector', 0);
```

## Event Serialization

### JSON Payload

`decimal.Decimal` is always stored as a **JSON string** (not number) to avoid float imprecision:

```json
// AccountOpened
{ "customer_id": "01hx...", "currency": "USD" }

// MoneyDeposited / MoneyWithdrawn
{ "amount": "100.50", "currency": "USD" }

// AccountFrozen
{ "reason": "KYC rejected" }

// AccountActivated  →  {} (empty payload)
```

### Event Registry

Each service maintains a **registry** mapping `event_type → deserializer`.
This is the only place that knows the concrete event types for that service.

```go
// internal/infrastructure/eventstore/registry.go
type EventFactory func(baseFields event.Base, payload []byte) (event.DomainEvent, error)

type Registry struct {
    factories map[string]EventFactory
}

func (r *Registry) Register(eventType string, f EventFactory)
func (r *Registry) Deserialize(eventType string, base event.Base, payload []byte) (event.DomainEvent, error)
```

`account_registry.go` wires all wallet domain event types at startup.

## Projector Binary

```
wallet-service/
  cmd/
    api/
      main.go          ← existing HTTP API
    projector/
      main.go          ← NEW: standalone projector process
```

The projector `main.go`:
1. Connects to `wallet_db`
2. Reads checkpoint (`last_global_seq`) from `projector_checkpoints`
3. Enters an event loop:
   - `SELECT * FROM events WHERE global_seq > $last ORDER BY global_seq ASC LIMIT 100`
   - Passes batch to `AccountProjector.Apply(ctx, tx, events)`
   - Updates `projector_checkpoints.last_global_seq`
   - If batch was empty → wait N ms (or use `pg_notify` wake-up), then retry

### Projector Interface

```go
// internal/infrastructure/projector/projector.go
type EventApplier interface {
    Apply(ctx context.Context, tx pgx.Tx, events []event.DomainEvent) error
}
```

### AccountProjector

```go
// internal/infrastructure/projector/account_projector.go
type AccountProjector struct{}

func (p *AccountProjector) Apply(ctx context.Context, tx pgx.Tx, events []event.DomainEvent) error {
    for _, e := range events {
        switch v := e.(type) {
        case domain.AccountOpened:
            // INSERT INTO account_read_models
        case domain.MoneyDeposited:
            // UPDATE account_read_models SET balance = balance + $1, version = $2
            // INSERT INTO transaction_history
        case domain.MoneyWithdrawn:
            // UPDATE account_read_models SET balance = balance - $1, version = $2
            // INSERT INTO transaction_history
        case domain.AccountActivated:
            // UPDATE account_read_models SET status = 'Active'
        case domain.AccountFrozen:
            // UPDATE account_read_models SET status = 'Frozen'
        }
    }
    return nil
}
```

## Updated Query Handlers

After this plan, query handlers read from **read model tables** — no event replay.
A new `ReadModelRepository` interface is introduced:

```go
// internal/infrastructure/readmodel/repository.go
type AccountReadRepository interface {
    GetBalance(ctx context.Context, accountID domain.AccountID) (appaccount.BalanceResult, error)
    GetTransactions(ctx context.Context, accountID domain.AccountID) ([]appaccount.TransactionRecord, error)
}
```

`GetBalanceHandler` and `GetTransactionsHandler` are updated to accept this interface
instead of `eventstore.EventStore`.

## Migration Strategy

Use **`github.com/pressly/goose/v3`** with plain SQL files in `db/migrations/`.
Migrations run automatically at startup (both `cmd/api` and `cmd/projector`).

```
wallet-service/db/migrations/
  001_create_events.sql
  002_create_read_models.sql
  003_create_projector_checkpoint.sql
```

## File Layout

```
wallet-service/
  cmd/
    api/main.go
    projector/main.go            ← NEW
  internal/
    infrastructure/
      eventstore/
        store.go                 (interface — unchanged)
        inmemory.go              (keep for unit tests)
        postgres.go              (NEW — PostgresEventStore)
        registry.go              (NEW — EventFactory, Registry)
        account_registry.go      (NEW — registers wallet domain events)
      projector/
        projector.go             (NEW — EventApplier interface + Runner)
        account_projector.go     (NEW — AccountProjector)
      readmodel/
        repository.go            (NEW — AccountReadRepository interface)
        postgres_repository.go   (NEW — SELECT from read model tables)
  db/
    migrations/
      001_create_events.sql
      002_create_read_models.sql
      003_create_projector_checkpoint.sql
```

`kyc-service` follows the same layout — `cmd/kyc-projector/`, own `kyc_db`,
own migrations, own projector for KYC read models.

## Acceptance Criteria

- [ ] Events survive service restart — state is fully restored from PostgreSQL
- [ ] Concurrent `Append` with same `expectedVersion` → one returns `ErrVersionConflict`
- [ ] `UNIQUE (aggregate_id, event_version)` enforced at DB level
- [ ] `Load` returns events ordered by `event_version ASC`
- [ ] Projector binary starts independently, reads from checkpoint, applies events
- [ ] After Deposit, `account_read_models.balance` updates (eventually)
- [ ] After deposit/withdraw, `transaction_history` has a new row
- [ ] Projector restart replays from last checkpoint — no duplicate rows
- [ ] `GetBalance` queries `account_read_models` — no event replay
- [ ] `GetTransactions` queries `transaction_history` — no event replay
- [ ] `decimal.Decimal` stored as `NUMERIC(20,8)`, amounts in JSONB as JSON string
- [ ] Unknown event type during deserialization → error, not panic
- [ ] `cmd/wallet-projector` and `cmd/wallet-api` are independent binaries

## Tasks

### Infrastructure
- [ ] Add PostgreSQL to `docker-compose.yml` (wallet_db + kyc_db as separate databases)
- [ ] Add `github.com/pressly/goose/v3` and `github.com/jackc/pgx/v5` dependencies
- [ ] Write SQL migrations (001, 002, 003)

### Event Store
- [ ] `eventstore/registry.go` — `EventFactory`, `Registry`
- [ ] `eventstore/account_registry.go` — register all 5 wallet domain event types
- [ ] `eventstore/postgres.go` — `PostgresEventStore` (Append + Load; no projector call)

### Projector
- [ ] `projector/projector.go` — `EventApplier` interface + `Runner` (poll loop + checkpoint update)
- [ ] `projector/account_projector.go` — `AccountProjector.Apply`
- [ ] `cmd/projector/main.go` — wire DB + registry + projector, start runner

### Read Model
- [ ] `readmodel/repository.go` — `AccountReadRepository` interface
- [ ] `readmodel/postgres_repository.go` — `GetBalance`, `GetTransactions`

### Application Layer
- [ ] Update `GetBalanceHandler` — accept `AccountReadRepository`, remove event replay
- [ ] Update `GetTransactionsHandler` — accept `AccountReadRepository`, remove event replay

### Wiring
- [ ] Update `cmd/api/main.go` — wire `PostgresEventStore` + `PostgresReadModelRepository` via fx
- [ ] Keep `InMemoryEventStore` available for unit tests

### Docs
- [ ] Update `docs/infrastructure/eventstore.md`
- [ ] Add `docs/infrastructure/projector.md` — async projector pattern, checkpoint, eventual consistency
