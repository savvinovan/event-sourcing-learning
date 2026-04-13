# Architecture Overview

## Summary

This project is a learning implementation of **Domain-Driven Design (DDD)** combined with **CQRS** (Command Query Responsibility Segregation) and **Event Sourcing** in Go.

The repository is a **monorepo** with three Go modules managed via `go.work`:

| Module | Path | Role |
|--------|------|------|
| `github.com/savvinovan/wallet-service` | `wallet-service/` | Wallet bounded context |
| `github.com/savvinovan/kyc-service` | `kyc-service/` | KYC bounded context |
| `github.com/savvinovan/event-sourcing-learning/contracts` | `contracts/` | Shared event schemas |

The architecture is organized into four layers with a strict dependency rule: outer layers depend on inner layers, never the reverse.

## Layer Diagram

```mermaid
graph TD
    HTTP["Interface Adapters\n(HTTP / chi)"]
    CMD["Application\nCommand Bus"]
    QRY["Application\nQuery Bus"]
    AGG["Domain\nAggregates"]
    EVT["Domain\nDomain Events"]
    ES["Infrastructure\nEvent Store"]

    HTTP --> CMD
    HTTP --> QRY
    CMD --> AGG
    QRY --> AGG
    AGG --> EVT
    CMD --> ES
```

## CQRS Flow

```mermaid
sequenceDiagram
    participant H as HTTP Handler
    participant CB as Command Bus
    participant CH as Command Handler
    participant A as Aggregate
    participant ES as Event Store

    H->>CB: Dispatch(cmd)
    CB->>CH: Handle(ctx, cmd)
    CH->>A: execute command
    A->>A: Record(DomainEvent)
    CH->>ES: Append(aggregateID, changes, version)
    CB-->>H: error | nil
```

## Event Sourcing Flow

```mermaid
sequenceDiagram
    participant CH as Command Handler
    participant ES as Event Store
    participant A as Aggregate

    CH->>ES: Load(aggregateID)
    ES-->>CH: []DomainEvent
    CH->>A: LoadFromHistory(events)
    Note over A: state rebuilt from events
    CH->>A: execute command
    A->>A: Record(newEvent)
    CH->>ES: Append(aggregateID, changes, expectedVersion)
    Note over ES: optimistic concurrency check on version
```

## Dependency Rule

```mermaid
graph TD
    interfaces --> application
    application --> domain
    infrastructure --> domain
```

The `domain` package has **zero external dependencies** — only the Go standard library.

## Architectural Decisions

See [ADR Index](decisions/README.md) for all accepted architectural decisions.

Key decisions:
- [ADR-001](decisions/adr-001-ddd-cqrs-es.md) — DDD + CQRS + Event Sourcing as core architecture
- [ADR-002](decisions/adr-002-tech-stack.md) — Technology stack
