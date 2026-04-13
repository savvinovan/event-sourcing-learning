# PLAN-002: Monorepo Restructuring with go.work

| | |
|-|-|
| **Status** | DONE |
| **Date** | 2026-04-13 |

## Goal

Reorganize the repository into a proper monorepo with two independent microservices
and a shared contracts module. Set up `go.work` for local development.

## Tasks

- [x] Create `contracts/` as a separate Go module (`github.com/savvinovan/event-sourcing-learning/contracts`)
- [x] Move current code into `wallet-service/` as a separate Go module (`github.com/savvinovan/wallet-service`)
- [x] Create `kyc-service/` skeleton as a separate Go module (`github.com/savvinovan/kyc-service`)
- [x] Initialize `go.work` at repo root, add all three modules
- [x] Add `replace` directives in `wallet-service/go.mod` and `kyc-service/go.mod` for `contracts`
- [x] Update all docs to reflect new structure

## Acceptance Criteria

- [x] `go work sync` runs without errors from repo root
- [x] `go build ./...` succeeds in each of: `wallet-service/`, `kyc-service/`, `contracts/`
- [x] `wallet-service` imports `contracts` module — verified by `go list -m all` showing contracts as dependency
- [x] `kyc-service` imports `contracts` module — same check
- [x] No `replace` directives point outside the repo
- [x] All existing docs updated — no references to old root-level module path

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
