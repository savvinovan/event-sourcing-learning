# ADR-001: DDD + CQRS + Event Sourcing as Core Architecture

| | |
|-|-|
| **Status** | Accepted |
| **Date** | 2026-04-13 |
| **RFC** | — |

## Context

This is a learning project focused on understanding event-driven architecture patterns in Go.
The goal is to build a working implementation that demonstrates the interplay between DDD, CQRS, and Event Sourcing in a typed, idiomatic Go codebase.

## Decision

Adopt **DDD + CQRS + Event Sourcing** as the core architectural approach:

- **DDD** — model business logic through aggregates, domain events, value objects, and bounded contexts.
- **CQRS** — separate the write side (commands mutate state) from the read side (queries read projections).
- **Event Sourcing** — aggregate state is derived entirely from a sequence of domain events. No mutable rows — only an append-only event log.

## Consequences

### Positive
- State changes are auditable by design — every mutation is a persisted event.
- Read models (projections) can be rebuilt from the event log at any time.
- Command and query models evolve independently.
- Aggregate behavior is fully unit-testable without a database.

### Negative / Trade-offs
- Higher initial complexity than a simple CRUD approach.
- Eventual consistency between write and read models requires careful handling.
- Querying current state requires either a projection or replaying the event log.
