# PLAN-002: Monorepo Restructuring with go.work

| | |
|-|-|
| **Status** | In Progress |
| **Date** | 2026-04-13 |

## Goal

Reorganize the repository into a proper monorepo with two independent microservices
and a shared contracts module. Set up `go.work` for local development.

## Tasks

- [ ] Create `contracts/` as a separate Go module (`github.com/savvinovan/event-sourcing-learning/contracts`)
- [ ] Move current code into `wallet-service/` as a separate Go module (`github.com/savvinovan/wallet-service`)
- [ ] Create `kyc-service/` skeleton as a separate Go module (`github.com/savvinovan/kyc-service`)
- [ ] Initialize `go.work` at repo root, add all three modules
- [ ] Add `replace` directives in `wallet-service/go.mod` and `kyc-service/go.mod` for `contracts`
- [ ] Update all docs to reflect new structure

## Acceptance Criteria

- [ ] `go work sync` runs without errors from repo root
- [ ] `go build ./...` succeeds in each of: `wallet-service/`, `kyc-service/`, `contracts/`
- [ ] `wallet-service` imports `contracts` module — verified by `go list -m all` showing contracts as dependency
- [ ] `kyc-service` imports `contracts` module — same check
- [ ] No `replace` directives point outside the repo
- [ ] All existing docs updated — no references to old root-level module path

## Expected Structure

```
event-sourcing-learning/
├── contracts/            # shared event schemas only, no business logic
│   ├── go.mod
│   └── events/
│       ├── kyc.go
│       └── wallet.go
├── wallet-service/       # moved from root
│   ├── go.mod
│   ├── cmd/api/
│   ├── internal/
│   └── ...
├── kyc-service/          # new
│   ├── go.mod
│   ├── cmd/api/
│   └── internal/
├── go.work
└── docs/
```
