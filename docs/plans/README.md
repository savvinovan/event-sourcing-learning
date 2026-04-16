# Plans (Roadmap)

## Overview

Plans capture intended development work. They are living documents — tasks are updated or added as work progresses.
Completed plans are marked `DONE` and preserved for historical reference.

Documentation and code may link to plans to indicate when a feature was introduced.

## Index

| ID | Title | Status | Date |
|----|-------|--------|------|
| [PLAN-001](plan-001-initial-setup.md) | Initial project setup — Go module, skeleton, docs-as-code | DONE | 2026-04-13 |
| [PLAN-002](plan-002-monorepo-restructure.md) | Monorepo restructuring with go.work | DONE | 2026-04-13 |
| [PLAN-003](plan-003-contracts-module.md) | Contracts module — shared event schemas | DONE | 2026-04-13 |
| [PLAN-004](plan-004-wallet-service.md) | Wallet Service — domain implementation | DONE | 2026-04-13 |
| [PLAN-005](plan-005-kyc-service.md) | KYC Service — domain implementation | DONE | 2026-04-13 |
| [PLAN-006](plan-006-event-driven-integration.md) | Event-driven integration via Kafka | DONE | 2026-04-15 |
| [PLAN-007](plan-007-postgresql-event-store.md) | PostgreSQL Event Store + Async Projector | DONE | 2026-04-14 |
| [PLAN-008](plan-008-integration-tests.md) | Integration tests with Testcontainers | DONE | 2026-04-15 |
| [PLAN-009](plan-009-event-versioning.md) | Event versioning and schema evolution (upcasting) | DONE | 2026-04-13 |
| [PLAN-010](plan-010-openapi-codegen.md) | OpenAPI-first HTTP layer with oapi-codegen + chi | DONE | 2026-04-15 |
| [PLAN-011](plan-011-frontend-scaffold.md) | Frontend scaffold + monorepo setup (Vite + React + TS) | Not Started | 2026-04-16 |
| [PLAN-012](plan-012-admin-api.md) | Admin backend API (accounts, events, projector status) | Not Started | 2026-04-16 |
| [PLAN-013](plan-013-jwt-auth.md) | JWT authentication — access + refresh tokens | Not Started | 2026-04-16 |
| [PLAN-014](plan-014-admin-frontend.md) | Admin frontend pages (dashboard, accounts, events) | Not Started | 2026-04-16 |

## Statuses

- **In Progress** — actively being worked on
- **DONE** — fully implemented
- **Not Started** — planned, not yet started
- **Cancelled** — dropped; reason noted in the plan file
