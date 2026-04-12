# ADR-002: Technology Stack

| | |
|-|-|
| **Status** | Accepted |
| **Date** | 2026-04-13 |
| **RFC** | — |

## Context

Need to select libraries for dependency injection, HTTP routing, configuration, and logging.
Choices must support testability, idiomatic Go, and minimal friction for a learning project.

## Decision

| Concern | Library | Rationale |
|---------|---------|-----------|
| Dependency Injection | `go.uber.org/fx` | Lifecycle management, constructor injection, module grouping; industry-standard |
| HTTP Router | `github.com/go-chi/chi/v5` | Lightweight, `net/http`-compatible, no magic, composable middleware |
| Configuration | `github.com/ilyakaznacheev/cleanenv` | Environment-variable-first, 12-factor compliant, struct tags |
| Logging | `log/slog` (stdlib) | Structured logging built into Go 1.21+, no external dependency needed |
| Go version | Go 1.25 | Latest stable at project inception |

## Consequences

- `uber/fx` requires providers to be functions returning concrete types or interfaces — encourages constructor injection.
- `chi` stays close to `net/http` stdlib — handlers are plain `http.HandlerFunc`, easy to test.
- `slog` requires no extra dependency and integrates naturally with `uber/fx` lifecycle logging.
- `cleanenv` reads `.env` files via `joho/godotenv` (transitive dep) — no extra setup needed for local development.
