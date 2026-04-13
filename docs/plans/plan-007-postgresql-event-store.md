# PLAN-007: PostgreSQL Event Store Implementation

| | |
|-|-|
| **Status** | Not Started |
| **Date** | 2026-04-13 |
| **Depends on** | [PLAN-004](plan-004-wallet-service.md), [PLAN-005](plan-005-kyc-service.md) |

## Goal

Replace the in-memory `EventStore` implementation with a real PostgreSQL-backed one.
Makes the learning project concrete — not just interfaces and abstractions,
but actual persistence you can inspect, query, and reason about.

## Schema Design

```sql
CREATE TABLE events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_id    TEXT        NOT NULL,
    aggregate_type  TEXT        NOT NULL,
    event_type      TEXT        NOT NULL,
    event_version   INT         NOT NULL,
    payload         JSONB       NOT NULL,
    occurred_at     TIMESTAMPTZ NOT NULL,
    UNIQUE (aggregate_id, event_version)
);

CREATE INDEX idx_events_aggregate_id ON events (aggregate_id, event_version ASC);
```

The `UNIQUE (aggregate_id, event_version)` constraint enforces optimistic concurrency at the DB level —
no two events for the same aggregate can have the same version.

## Acceptance Criteria

- [ ] Events survive service restart — state is fully restored from PostgreSQL on startup
- [ ] Concurrent `Append` with the same `expectedVersion` — one succeeds, one returns `ErrVersionConflict`
- [ ] `UNIQUE (aggregate_id, event_version)` constraint is enforced at DB level (verified by direct SQL insert attempt)
- [ ] `Load` returns events in correct version order
- [ ] Old events with missing fields deserialize without panic (default zero values)
- [ ] `go test ./...` passes against a real PostgreSQL instance (via testcontainers — see PLAN-008)

## Tasks

- [ ] Add PostgreSQL to `docker-compose.yml`
- [ ] Write SQL migration (using `golang-migrate` or plain SQL files)
- [ ] Implement `PostgresEventStore` in `wallet-service/internal/infrastructure/eventstore/postgres.go`
- [ ] Implement `PostgresEventStore` in `kyc-service/internal/infrastructure/eventstore/postgres.go`
- [ ] Serialization: domain events → JSONB payload (JSON marshaling with event type as discriminator)
- [ ] Deserialization: JSONB payload → concrete event structs (event registry / type switch)
- [ ] Wire `PostgresEventStore` into uber/fx instead of in-memory store
- [ ] Update docs
