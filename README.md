# Event Sourcing Learning

A learning implementation of **DDD + CQRS + Event Sourcing** in Go.

Two microservices communicate via async domain events. Each service owns its database,
stores state as an immutable event stream, and maintains read models via a dedicated
projector process.

## What's inside

| Service | Description |
|---------|-------------|
| `wallet-service` | Accounts, deposits, withdrawals. HTTP API + async projector. |
| `kyc-service` | KYC verification workflow. Publishes `KYCVerified` / `KYCRejected` events. |
| `contracts` | Shared Kafka event schemas between services. |

**Key patterns used:**
- Event Sourcing — state derived entirely from an immutable event log
- CQRS — separate write path (command bus → aggregate → event store) and read path (projector → read model tables → query handlers)
- Async projection — `cmd/projector` tails the `events` table by `global_seq` and maintains `account_read_models` + `transaction_history`
- Optimistic concurrency — `UNIQUE(aggregate_id, event_version)` in PostgreSQL
- Typed domain IDs (`AccountID`, `CustomerID`) — never plain `string`
- Money value object (`decimal.Decimal` + ISO 4217 currency, never `float64`)

## Quick start

### Prerequisites

- Go 1.25+
- Docker + Docker Compose

### 1. Start databases

```bash
docker compose up -d
```

Starts `wallet-db` on port `5432` and `kyc-db` on port `5433`.

### 2. Run wallet API

```bash
cd wallet-service
go run ./cmd/api
```

Migrations are applied automatically on startup.
API listens on `http://localhost:8080`.

### 3. Run wallet projector (separate terminal)

```bash
cd wallet-service
go run ./cmd/projector
```

Tails the `events` table and updates read models every 500 ms.

### 4. Run KYC API (optional)

```bash
cd kyc-service
go run ./cmd/api
```

KYC API listens on `http://localhost:8081`.

## Configuration

Both services are configured via environment variables (12-factor).

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_DSN` | `postgres://wallet:wallet@localhost:5432/wallet?sslmode=disable` | PostgreSQL DSN |
| `HTTP_HOST` | `0.0.0.0` | HTTP listen address |
| `HTTP_PORT` | `8080` | HTTP listen port |
| `LOG_LEVEL` | `info` | Log level (`debug`, `info`, `warn`, `error`) |

## API endpoints (wallet-service)

```
POST   /accounts                    Open a new account
POST   /accounts/{id}/deposit       Deposit money
POST   /accounts/{id}/withdraw      Withdraw money
POST   /accounts/{id}/activate      Activate account (KYC approved)
POST   /accounts/{id}/freeze        Freeze account (KYC rejected)
GET    /accounts/{id}/balance       Get current balance
GET    /accounts/{id}/transactions  Get transaction history
GET    /healthz                     Health check
```

## Repository structure

```
event-sourcing-learning/
├── contracts/          # Shared event schemas (go module)
├── wallet-service/
│   ├── cmd/
│   │   ├── api/        # HTTP API binary
│   │   └── projector/  # Async projector binary
│   ├── db/migrations/  # goose SQL migrations
│   ├── internal/
│   │   ├── app/                    # fx wiring
│   │   ├── application/account/    # CQRS commands, queries, handlers
│   │   ├── domain/account/         # Aggregate, events, value objects
│   │   └── infrastructure/
│   │       ├── eventstore/         # PostgresEventStore + event registry
│   │       ├── projector/          # Async projector runner + AccountProjector
│   │       └── readmodel/          # Read model repository
│   └── go.mod
├── kyc-service/        # KYC bounded context (same structure)
├── docker-compose.yml  # wallet-db + kyc-db
├── go.work             # Go workspace
└── docs/               # Architecture docs, ADRs, plans
```

## Documentation

Full architecture docs, ADRs, and implementation plans live in [`docs/`](docs/README.md).
