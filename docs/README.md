# Event Sourcing Learning — Documentation

Project documentation following the **docs-as-code** approach.
All content reflects the current state of the codebase.

## Navigation

### Architecture
- [Architecture Overview](architecture/overview.md) — DDD, CQRS, Event Sourcing, layer diagram
- [Architecture Decision Records (ADR)](architecture/decisions/README.md) — accepted architectural decisions
- [Request for Comments (RFC)](architecture/rfcs/README.md) — proposals, under review, and rejected ideas

### Domain Layer
- [Domain Overview](domain/README.md) — bounded contexts, aggregates, events
- [Aggregate Root](domain/shared/aggregate.md) — base aggregate with event sourcing support
- [Domain Events](domain/shared/event.md) — event interface and base implementation

### Application Layer
- [Application Overview](application/README.md) — CQRS command and query sides
- [Command Bus](application/commands.md) — Command, Handler, Bus interfaces
- [Query Bus](application/queries.md) — Query, Handler, Bus interfaces

### Infrastructure Layer
- [Infrastructure Overview](infrastructure/README.md)
- [Event Store](infrastructure/eventstore.md) — persistence interface with optimistic concurrency

### Interface Adapters
- [Interfaces Overview](interfaces/README.md)
- [HTTP API](interfaces/http.md) — chi router, middleware, endpoints

### Plans
- [Roadmap](plans/README.md) — past and future development plans
