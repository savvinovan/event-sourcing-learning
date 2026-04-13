# PLAN-004: Wallet Service — Domain Implementation

| | |
|-|-|
| **Status** | Completed |
| **Date** | 2026-04-13 |
| **Depends on** | [PLAN-002](plan-002-monorepo-restructure.md), [PLAN-003](plan-003-contracts-module.md) |

## Goal

Implement the Wallet bounded context with full DDD + CQRS + Event Sourcing.
Accounts can be opened, funded, withdrawn from, and transferred between.
Transactions are blocked until KYC is verified.

## Aggregate: Account

### Commands
- `OpenAccount` — create a new account for a customer
- `DepositMoney` — credit funds to account
- `WithdrawMoney` — debit funds (must check balance, must be KYC verified)
- `InitiateTransfer` — start transfer to another account
- `ActivateAccount` — triggered by `KYCVerified` event from kyc-service
- `FreezeAccount` — triggered by `KYCRejected` event from kyc-service

### Domain Events
- `AccountOpened`
- `MoneyDeposited`
- `MoneyWithdrawn`
- `TransferInitiated`
- `AccountActivated`
- `AccountFrozen`

### Business Rules
- Withdrawal requires sufficient balance
- Withdrawal and transfer require account to be in `Active` status
- New account starts in `Pending` status (awaiting KYC)
- Account becomes `Active` on `AccountActivated`
- Account becomes `Frozen` on `AccountFrozen`

## Read Models (Projections)

- **Account Balance** — current balance per account
- **Transaction History** — list of all deposits/withdrawals/transfers for an account

## HTTP Endpoints

- `POST /accounts` — open account
- `POST /accounts/{id}/deposit` — deposit money
- `POST /accounts/{id}/withdraw` — withdraw money
- `POST /accounts/{id}/transfer` — initiate transfer
- `GET /accounts/{id}/balance` — get current balance
- `GET /accounts/{id}/transactions` — get transaction history

## Acceptance Criteria

- [x] `POST /accounts` creates an account in `Pending` status — verified via `GET /accounts/{id}/balance`
- [x] `POST /accounts/{id}/deposit` increases balance — verified via balance endpoint
- [x] `POST /accounts/{id}/withdraw` on a `Pending` account returns error (not KYC verified)
- [x] `POST /accounts/{id}/withdraw` on an `Active` account with sufficient balance succeeds
- [x] `POST /accounts/{id}/withdraw` with insufficient balance returns error
- [x] Two concurrent withdrawals that together exceed balance — only one succeeds (optimistic concurrency)
- [x] `GET /accounts/{id}/transactions` returns all deposits and withdrawals in order
- [x] Account transitions to `Active` when `AccountActivated` command is dispatched
- [x] Account transitions to `Frozen` when `AccountFrozen` command is dispatched
- [x] All account state is fully reconstructible by replaying events from the event store

## Tasks

- [x] `Account` aggregate with all commands and events
- [x] `AccountStatus` value object (`Pending`, `Active`, `Frozen`)
- [x] `Money` value object (amount + currency)
- [x] Command handlers for all write operations
- [x] In-memory event store implementation
- [x] Balance projection (query handler)
- [x] Transaction history projection (query handler)
- [x] HTTP handlers and router wiring
- [x] uber/fx module for wallet domain
- [ ] Subscribe to KYC events from kyc-service (see [PLAN-006](plan-006-event-driven-integration.md))
- [x] Update docs

> **Note:** `InitiateTransfer` command/event and `POST /accounts/{id}/transfer` endpoint are deferred — no acceptance criteria require them and they depend on PLAN-005 (KYC) and PLAN-006 (Kafka).
