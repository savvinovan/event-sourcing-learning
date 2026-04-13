# PLAN-003: Contracts Module — Shared Event Schemas

| | |
|-|-|
| **Status** | Not Started |
| **Date** | 2026-04-13 |
| **Depends on** | [PLAN-002](plan-002-monorepo-restructure.md) |

## Goal

Define the shared event contracts between `wallet-service` and `kyc-service`.
The `contracts` module contains only event structs — no business logic, no dependencies on domain code.
This gives compile-time safety when either service changes an event schema.

## Tasks

- [ ] Define KYC events in `contracts/events/kyc.go`
  - `KYCSubmitted`
  - `KYCVerified`
  - `KYCRejected`
- [ ] Define Wallet events in `contracts/events/wallet.go`
  - `WalletActivated`
  - `WalletFrozen`
- [ ] Add docs: `docs/contracts/README.md` — explain why contracts module exists and what belongs here
- [ ] Add docs: `docs/contracts/events.md` — document each event schema

## Acceptance Criteria

- [ ] `contracts/` has zero external dependencies (`go.mod` shows only stdlib)
- [ ] All KYC event structs compile and are importable from both services
- [ ] All Wallet event structs compile and are importable from both services
- [ ] Changing a field in any event struct causes a compile error in the service that uses it
- [ ] Docs cover every exported event struct with field descriptions

## Rules for this module

- Zero external dependencies (only stdlib)
- Only struct definitions and constants
- No methods with business logic
- Changing any struct here is a breaking change — both services must be updated
