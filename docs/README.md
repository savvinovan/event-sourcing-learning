# Event Sourcing Learning — Documentation

Project documentation following the **docs-as-code** approach.
All content reflects the current state of the codebase.

## Repository Structure

```
event-sourcing-learning/
├── contracts/        # Shared event schemas (github.com/savvinovan/event-sourcing-learning/contracts)
├── wallet-service/   # Wallet microservice (github.com/savvinovan/wallet-service)
├── kyc-service/      # KYC microservice (github.com/savvinovan/kyc-service)
├── go.work           # Workspace — local module resolution
└── docs/             # This documentation
```

## Navigation

### Architecture
- [Architecture Overview](architecture/overview.md) — DDD, CQRS, Event Sourcing, layer diagram
- [Architecture Decision Records (ADR)](architecture/decisions/README.md) — accepted architectural decisions
- [Request for Comments (RFC)](architecture/rfcs/README.md) — proposals, under review, and rejected ideas

### Contracts
- [Contracts Overview](contracts/README.md) — shared event schemas between services

### Wallet Service
- [Domain Overview](domain/README.md) — bounded contexts, aggregates, events
- [Aggregate Root](domain/shared/aggregate.md) — base aggregate with event sourcing support
- [Domain Events](domain/shared/event.md) — event interface and base implementation
- [Application Overview](application/README.md) — CQRS command and query sides
- [Command Bus](application/commands.md) — Command, Handler, Bus interfaces
- [Query Bus](application/queries.md) — Query, Handler, Bus interfaces
- [Infrastructure Overview](infrastructure/README.md)
- [Event Store](infrastructure/eventstore.md) — persistence interface with optimistic concurrency
- [Interfaces Overview](interfaces/README.md)
- [HTTP API](interfaces/http.md) — chi router, middleware, endpoints

### Plans
- [Roadmap](plans/README.md) — past and future development plans
