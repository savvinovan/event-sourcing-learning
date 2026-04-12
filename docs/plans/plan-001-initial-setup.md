# PLAN-001: Initial Project Setup

| | |
|-|-|
| **Status** | DONE |
| **Date** | 2026-04-13 |

## Goal

Bootstrap the Go project with the full skeleton for DDD + CQRS + Event Sourcing,
establish docs-as-code documentation structure, and push to GitHub.

## Tasks

- [x] Initialize Go module `github.com/savvinovan/event-sourcing-learning`
- [x] Add dependencies: `chi`, `uber/fx`, `cleanenv`
- [x] Domain layer — `aggregate.Root`, `event.DomainEvent`, `event.Base`
- [x] Application layer — `command.CommandType`, `command.Bus`, `command.Handler`, `query.Bus`, `query.Handler` interfaces
- [x] Infrastructure layer — `eventstore.EventStore` interface with `ErrVersionConflict`
- [x] Interface adapters — chi router, `GET /health` endpoint
- [x] Entry point — `cmd/api/main.go` with uber/fx lifecycle wiring
- [x] Config — `config.Config` via cleanenv (HTTP + Log)
- [x] Makefile (run / build / test / lint / tidy)
- [x] `.gitignore`
- [x] Push to `git@github.com:savvinovan/event-sourcing-learning.git`
- [x] Docs-as-code structure in `docs/` with Markdown + Mermaid

## Implemented In

- `internal/domain/aggregate/aggregate.go`
- `internal/domain/event/event.go`
- `internal/application/command/bus.go`
- `internal/application/query/bus.go`
- `internal/infrastructure/eventstore/store.go`
- `internal/interfaces/http/router.go`
- `internal/interfaces/http/handler/health.go`
- `cmd/api/main.go`
- `config/config.go`
