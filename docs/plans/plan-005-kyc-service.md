# PLAN-005: KYC Service — Domain Implementation

| | |
|-|-|
| **Status** | Not Started |
| **Date** | 2026-04-13 |
| **Depends on** | [PLAN-002](plan-002-monorepo-restructure.md), [PLAN-003](plan-003-contracts-module.md) |

## Goal

Implement the KYC bounded context as an independent microservice.
Handles identity verification lifecycle for customers.
Publishes events consumed by `wallet-service` to activate or freeze accounts.

## Aggregate: KYCVerification

### Commands
- `SubmitKYC` — customer submits identity documents
- `ApproveKYC` — operator approves the verification
- `RejectKYC` — operator rejects the verification

### Domain Events
- `KYCSubmitted`
- `KYCVerified`
- `KYCRejected`

### Status State Machine

```mermaid
stateDiagram-v2
    [*] --> Submitted: SubmitKYC
    Submitted --> Verified: ApproveKYC
    Submitted --> Rejected: RejectKYC
    Verified --> [*]
    Rejected --> [*]
```

### Business Rules
- Only `Submitted` verification can be approved or rejected
- Once verified or rejected, status is final (no transitions out)
- One KYC verification per customer at a time

## Read Models (Projections)

- **KYC Status** — current verification status per customer

## HTTP Endpoints

- `POST /kyc` — submit KYC for a customer
- `POST /kyc/{id}/approve` — approve verification (operator)
- `POST /kyc/{id}/reject` — reject verification (operator)
- `GET /kyc/{id}` — get current KYC status

## Acceptance Criteria

- [ ] `POST /kyc` creates a verification in `Submitted` status — verified via `GET /kyc/{id}`
- [ ] `POST /kyc/{id}/approve` transitions status to `Verified`
- [ ] `POST /kyc/{id}/reject` transitions status to `Rejected`
- [ ] Approving an already `Verified` verification returns an error
- [ ] Rejecting an already `Rejected` verification returns an error
- [ ] `KYCVerified` event is published to Kafka after approval (verified in integration test)
- [ ] `KYCRejected` event is published to Kafka after rejection (verified in integration test)
- [ ] All KYC state is fully reconstructible by replaying events from the event store

## Tasks

- [ ] `KYCVerification` aggregate with all commands and events
- [ ] `KYCStatus` value object (`Submitted`, `Verified`, `Rejected`)
- [ ] Command handlers
- [ ] In-memory event store implementation
- [ ] KYC status projection (query handler)
- [ ] HTTP handlers and router wiring
- [ ] uber/fx module for KYC domain
- [ ] Publish `KYCVerified` / `KYCRejected` to message broker (see [PLAN-006](plan-006-event-driven-integration.md))
- [ ] Update docs
