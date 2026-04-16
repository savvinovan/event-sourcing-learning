# PLAN-012: Admin Backend API

| | |
|-|-|
| **Status** | Not Started |
| **Date** | 2026-04-16 |
| **Depends on** | [PLAN-010](plan-010-openapi-codegen.md) |

## Goal

Add a read-only admin HTTP API to wallet-service under `/admin/` that exposes internal state
for the admin panel: account list, raw event streams, projector checkpoint, and a dashboard summary.

Auth middleware (JWT) is added in [PLAN-013](plan-013-jwt-auth.md) — for now routes are open.
The API is defined OpenAPI-first (same oapi-codegen pattern as PLAN-010).

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/admin/dashboard` | Summary stats: accounts by status, total events, projector lag |
| `GET` | `/admin/accounts` | Paginated list of all accounts (from read model) |
| `GET` | `/admin/accounts/{id}` | Single account detail (balance, status, customer ID) |
| `GET` | `/admin/accounts/{id}/events` | Raw event stream for an account (from event store) |
| `GET` | `/admin/projector/status` | Projector checkpoint + global_seq lag |

## Request / Response Shapes

### `GET /admin/accounts?page=1&limit=50`

```json
{
  "items": [
    {
      "account_id": "...",
      "customer_id": "...",
      "status": "Active",
      "balance": "150.00",
      "currency": "USD",
      "version": 5
    }
  ],
  "total": 142,
  "page": 1,
  "limit": 50
}
```

### `GET /admin/accounts/{id}/events`

```json
{
  "account_id": "...",
  "events": [
    {
      "event_type": "AccountOpened",
      "event_version": 1,
      "schema_version": 1,
      "occurred_at": "2026-04-16T10:00:00Z",
      "payload": { "customer_id": "...", "currency": "USD" }
    },
    {
      "event_type": "MoneyDeposited",
      "event_version": 2,
      "schema_version": 2,
      "occurred_at": "2026-04-16T10:05:00Z",
      "payload": { "amount": "100.00", "currency": "USD", "description": "" }
    }
  ]
}
```

Payload is returned as raw `json.RawMessage` — no domain deserialization needed here,
just read directly from the `events` table.

### `GET /admin/projector/status`

```json
{
  "projector_name": "account_projector",
  "last_processed_seq": 1042,
  "latest_global_seq": 1045,
  "lag": 3
}
```

### `GET /admin/dashboard`

```json
{
  "accounts": {
    "total": 142,
    "by_status": {
      "Pending": 12,
      "Active": 125,
      "Frozen": 5
    }
  },
  "events": {
    "total": 1045
  },
  "projector": {
    "lag": 3
  }
}
```

## Implementation

### OpenAPI spec

New file: `wallet-service/api/admin-openapi.yaml`
New codegen config: `wallet-service/api/admin-oapi-codegen.yaml`
Generated output: `wallet-service/internal/interfaces/http/gen/admin.gen.go`

### Handler

`wallet-service/internal/interfaces/http/handler/admin.go`

```go
type AdminHandler struct {
    db *pgxpool.Pool // direct DB reads for admin queries
}
```

Admin reads bypass the command/query bus — they query DB directly (read model + events table).
This is intentional: admin is a reporting concern, not a domain concern.

### Router

New route group on the existing chi router:

```go
r.Route("/admin", func(r chi.Router) {
    // PLAN-013 will add: r.Use(jwtAdminMiddleware)
    admin.HandlerFromMux(adminStrictHandler, r)
})
```

## Acceptance Criteria

- [ ] `GET /admin/dashboard` returns correct counts from DB
- [ ] `GET /admin/accounts` returns paginated accounts from `account_read_models`
- [ ] `GET /admin/accounts/{id}/events` returns raw event rows as JSON (payload as-is)
- [ ] `GET /admin/projector/status` returns checkpoint + lag
- [ ] `go generate ./api/...` regenerates admin code without errors
- [ ] `go build ./...` passes
- [ ] All endpoints return `404` for unknown account IDs
- [ ] Pagination `page`/`limit` params validated (limit max 200)

## Tasks

- [ ] Write `wallet-service/api/admin-openapi.yaml` with all 5 endpoint specs
- [ ] Add `wallet-service/api/admin-oapi-codegen.yaml` codegen config
- [ ] Run `go generate` → commit `admin.gen.go`
- [ ] Implement `AdminHandler` with direct DB queries
- [ ] Wire `/admin` route group in `router.go`
- [ ] Wire `AdminHandler` in `providers.go` / Fx app
- [ ] Update docs
