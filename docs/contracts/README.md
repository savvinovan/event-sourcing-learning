# Contracts Module

**Module:** `github.com/savvinovan/event-sourcing-learning/contracts`
**Source:** `contracts/`

## Purpose

The contracts module contains **only shared event schemas** — the cross-service communication contracts.
It has zero external dependencies and no business logic.

Both `wallet-service` and `kyc-service` import it. Changing any event struct here
causes a compile error in every service that uses it — giving compile-time safety
against schema drift between services.

## Contents

- `events/kyc.go` — KYC events published by `kyc-service`
- `events/wallet.go` — Wallet events published by `wallet-service`

## What Belongs Here

- Event struct definitions
- Topic name constants (Kafka topic strings)

## What Does NOT Belong Here

- Business logic
- Repository interfaces
- Domain types or value objects
- Any code with external dependencies

## Event Schemas

### KYC Events (`events/kyc.go`)

| Struct | Published By | Consumed By | Trigger |
|--------|-------------|-------------|---------|
| `KYCSubmitted` | kyc-service | — | Customer submits documents |
| `KYCVerified` | kyc-service | wallet-service | Operator approves KYC |
| `KYCRejected` | kyc-service | wallet-service | Operator rejects KYC |

### Wallet Events (`events/wallet.go`)

| Struct | Published By | Consumed By | Trigger |
|--------|-------------|-------------|---------|
| `WalletActivated` | wallet-service | — | Account activated after KYC verified |
| `WalletFrozen` | wallet-service | — | Account frozen after KYC rejected |

## See Also

- [Architecture Overview](../architecture/overview.md)
- [PLAN-003](../plans/plan-003-contracts-module.md) — full contracts implementation plan
