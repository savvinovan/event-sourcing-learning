# Async Projector

**Source:** `wallet-service/internal/infrastructure/projector/` and `wallet-service/cmd/projector/`

## Overview

The async projector is a **separate binary** (`cmd/projector`) that tails the `events` table and builds read models in PostgreSQL. It runs independently from the API server.

Design principles:
- **Poll-based** — polls the `events` table using `global_seq BIGSERIAL` as a cursor
- **Single-instance** — uses `SELECT ... FOR UPDATE` on the checkpoint row to prevent concurrent runs
- **Transactional** — event processing and checkpoint update happen in a single DB transaction

## Architecture

```mermaid
graph TD
    API["cmd/api\n(wallet API server)"]
    PROJ["cmd/projector\n(async projector)"]
    DB[("PostgreSQL")]

    API -->|"Append events\n(events table)"| DB
    PROJ -->|"Poll events\n(global_seq cursor)"| DB
    PROJ -->|"Update read models\n(account_read_models,\ntransaction_history)"| DB
    PROJ -->|"Update checkpoint\n(projector_checkpoints)"| DB
```

## Runner

**File:** `projector.go`

```go
type Runner struct {
    db       *pgxpool.Pool
    registry *eventstore.Registry
    applier  EventApplier
    // config: batchSize, pollInterval
}

func (r *Runner) Run(ctx context.Context) error
```

`Run` loops until `ctx` is cancelled:
1. Call `processBatch` — returns `(processed int, err error)`
2. If `processed == batchSize` → re-poll immediately (more events likely pending)
3. Otherwise → sleep `pollInterval` (default 500ms)

### processBatch

```
BEGIN
SELECT last_seq FROM projector_checkpoints WHERE name = 'account_projector' FOR UPDATE
SELECT * FROM events WHERE global_seq > last_seq ORDER BY global_seq LIMIT batchSize
FOR each event:
    deserialize via Registry
    call EventApplier.Apply(ctx, tx, event)
UPDATE projector_checkpoints SET last_seq = maxProcessed WHERE name = 'account_projector'
COMMIT
```

`FOR UPDATE` on the checkpoint row ensures only one projector instance processes events at a time.

## EventApplier Interface

```go
type EventApplier interface {
    Apply(ctx context.Context, tx pgx.Tx, e event.DomainEvent) error
}
```

## AccountProjector

**File:** `account_projector.go`

Implements `EventApplier`. Uses a handler-registry pattern — no switch statement:

```go
type AccountProjector struct {
    handlers map[string]eventHandler
}

type eventHandler func(ctx context.Context, tx pgx.Tx, e event.DomainEvent) error
```

Handlers registered in `NewAccountProjector()`:

| Event | Handler | Read Model Updated |
|-------|---------|-------------------|
| `AccountOpened` | `onAccountOpened` | `account_read_models` (INSERT) |
| `MoneyDeposited` | `onMoneyDeposited` | `account_read_models` (balance +), `transaction_history` (INSERT) |
| `MoneyWithdrawn` | `onMoneyWithdrawn` | `account_read_models` (balance −), `transaction_history` (INSERT) |
| `AccountActivated` | `onAccountActivated` | `account_read_models` (status = Active) |
| `AccountFrozen` | `onAccountFrozen` | `account_read_models` (status = Frozen) |

## Read Model Schema

```sql
-- Account state (balance + status)
CREATE TABLE account_read_models (
    account_id  UUID PRIMARY KEY,
    customer_id UUID NOT NULL,
    currency    TEXT NOT NULL,
    balance     NUMERIC(20,8) NOT NULL DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'Pending'
);

-- Transaction history
CREATE TABLE transaction_history (
    id          UUID PRIMARY KEY,
    account_id  UUID NOT NULL,
    tx_type     TEXT NOT NULL,  -- 'deposit' | 'withdrawal'
    amount      NUMERIC(20,8) NOT NULL,
    currency    TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL
);

-- Projector cursor
CREATE TABLE projector_checkpoints (
    name     TEXT PRIMARY KEY,
    last_seq BIGINT NOT NULL DEFAULT 0
);
-- Seeded with: INSERT INTO projector_checkpoints (name, last_seq) VALUES ('account_projector', 0)
```

`NUMERIC(20,8)` for all monetary amounts — never `float`. Read back as `TEXT` via `::text` cast, parsed with `decimal.NewFromString()`.

## Checkpoint

The `projector_checkpoints` table holds one row per projector (`account_projector`). `last_seq` is the highest `global_seq` processed. On restart the projector resumes from `last_seq + 1`.

## Running the Projector

```bash
# Set DB connection string
export DSN=postgres://wallet:wallet@localhost:5432/wallet?sslmode=disable

# Run (projector applies migrations before starting)
go run ./cmd/projector/...
```

The projector binary:
1. Loads config from environment
2. Creates `pgxpool`
3. Runs goose migrations (same embedded migrations as the API binary)
4. Starts the `Runner` event loop
5. Shuts down gracefully on `SIGTERM` / `SIGINT`

## See Also

- [Event Store](eventstore.md) — writes events that the projector consumes
- [PLAN-007](../plans/plan-007-postgresql-event-store.md) — implementation plan
