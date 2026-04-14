# PLAN-010: OpenAPI-First HTTP Layer with oapi-codegen

| | |
|-|-|
| **Status** | DONE |
| **Date** | 2026-04-13 |
| **Depends on** | [PLAN-004](plan-004-wallet-service.md), [PLAN-005](plan-005-kyc-service.md) |

## Goal

Replace hand-written HTTP handlers and DTOs with code generated from OpenAPI 3.0 specifications.
Each service defines its contract as an `api/openapi.yaml` spec; `oapi-codegen` generates the
chi-compatible server interface and request/response models. Developers only implement the
`StrictServerInterface` — no manual JSON decoding, no route registration boilerplate.

## Approach

**Tool:** [`github.com/oapi-codegen/oapi-codegen`](https://github.com/oapi-codegen/oapi-codegen) v2

**Mode:** `strict-server` + `chi-server` — handlers receive typed request structs and return typed
response structs; the generated layer handles routing, decoding, validation, and encoding.

**Flow:**
```
api/openapi.yaml
       ↓  go generate
internal/interfaces/http/gen/
    server.gen.go   ← StrictServerInterface + chi routes
    models.gen.go   ← request/response types
       ↓  implement
internal/interfaces/http/handler/
    account.go      ← implements StrictServerInterface
```

## Per-Service Changes

### wallet-service

**Files to add:**
- `api/openapi.yaml` — OpenAPI 3.0 spec for all account endpoints
- `api/oapi-codegen.yaml` — generator config (strict chi server + models)
- `internal/interfaces/http/gen/` — generated code (committed, not gitignored)

**Files to remove:**
- `internal/interfaces/http/handler/dto.go` — replaced by generated models
- Manual route registration in `router.go` — replaced by `HandlerFromMux`

**Files to update:**
- `internal/interfaces/http/handler/account.go` — implement `StrictServerInterface` instead of raw `http.Handler` methods
- `internal/interfaces/http/router.go` — wire generated `HandlerFromMux`

### kyc-service

Same structure:
- `api/openapi.yaml`, `api/oapi-codegen.yaml`
- `internal/interfaces/http/gen/`
- Update `handler/kyc.go` to implement `StrictServerInterface`
- Update `router.go`

## Generator Config

`api/oapi-codegen.yaml`:
```yaml
package: gen
output: ../internal/interfaces/http/gen/server.gen.go
generate:
  strict-server: true
  chi-server: true
  models: true
  embedded-spec: false
```

## go:generate Directive

In each service's `api/` directory, add `doc.go`:
```go
package api

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config oapi-codegen.yaml openapi.yaml
```

Run with: `go generate ./api/...`

## StrictServerInterface Pattern

Generated interface (example for wallet-service):
```go
type StrictServerInterface interface {
    OpenAccount(ctx context.Context, request OpenAccountRequestObject) (OpenAccountResponseObject, error)
    Deposit(ctx context.Context, request DepositRequestObject) (DepositResponseObject, error)
    Withdraw(ctx context.Context, request WithdrawRequestObject) (WithdrawResponseObject, error)
    GetBalance(ctx context.Context, request GetBalanceRequestObject) (GetBalanceResponseObject, error)
    GetTransactions(ctx context.Context, request GetTransactionsRequestObject) (GetTransactionsResponseObject, error)
}
```

Handler implementation dispatches to command/query bus — same logic, cleaner signature:
```go
func (h *AccountHandler) OpenAccount(ctx context.Context, req gen.OpenAccountRequestObject) (gen.OpenAccountResponseObject, error) {
    accountID := domain.NewAccountID()
    cmd := appaccount.OpenAccountCommand{
        AccountID:  accountID,
        CustomerID: domain.CustomerID(req.Body.CustomerId),
        Currency:   req.Body.Currency,
    }
    if err := h.commands.Dispatch(ctx, cmd); err != nil {
        return mapCommandError(err), nil
    }
    return gen.OpenAccount201JSONResponse{AccountId: string(accountID)}, nil
}
```

## Acceptance Criteria

- [ ] `wallet-service/api/openapi.yaml` defines all 5 account endpoints with request/response schemas
- [ ] `kyc-service/api/openapi.yaml` defines all 4 KYC endpoints with request/response schemas
- [ ] `go generate ./api/...` runs without errors in both services
- [ ] Generated code compiles — `go build ./...` passes
- [ ] `AccountHandler` implements the generated `StrictServerInterface` (compile-time check)
- [ ] `KYCHandler` implements the generated `StrictServerInterface` (compile-time check)
- [ ] Hand-written DTOs (`dto.go`) deleted in both services
- [ ] All existing HTTP endpoints still reachable and return correct responses
- [ ] `GET /health` still works (not generated — keep as-is)
- [ ] `oapi-codegen` added as a tool dependency in each service's `go.mod` (`go tool`)

## Tasks

- [ ] Write `wallet-service/api/openapi.yaml`
- [ ] Write `kyc-service/api/openapi.yaml`
- [ ] Add `oapi-codegen` as `go tool` to both `go.mod`
- [ ] Add `api/oapi-codegen.yaml` configs to both services
- [ ] Add `api/doc.go` with `//go:generate` directive to both services
- [ ] Run `go generate`, commit generated files
- [ ] Rewrite `wallet-service` handler to implement `StrictServerInterface`
- [ ] Rewrite `kyc-service` handler to implement `StrictServerInterface`
- [ ] Update `router.go` in both services to use `HandlerFromMux`
- [ ] Delete `dto.go` from both services
- [ ] Update docs
