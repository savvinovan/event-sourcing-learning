# Application Layer

## Overview

The application layer implements **CQRS** — it splits the write path (commands) from the read path (queries).
It orchestrates domain objects and infrastructure but contains no business rules itself.

```mermaid
graph TD
    HTTP[HTTP Handler]
    CB[command.Bus]
    QB[query.Bus]
    CH[command.Handler]
    QH[query.Handler]
    AGG[Aggregate]

    HTTP -->|Dispatch| CB
    HTTP -->|Ask| QB
    CB --> CH
    QB --> QH
    CH --> AGG
    QH --> AGG
```

## Contents

- [Command Bus](commands.md) — `internal/application/command/bus.go`
- [Query Bus](queries.md) — `internal/application/query/bus.go`
