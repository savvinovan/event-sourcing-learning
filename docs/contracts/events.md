# Contract Event Schemas

All event structs live in `contracts/events/`. Topic name constants live in `contracts/topics/`.

Changing any field here is a **breaking change** — both publisher and consumer must be updated simultaneously.
The compiler enforces this: services reference each field explicitly, so removing or renaming a field causes a compile error.

---

## KYC Events (`contracts/events/kyc.go`)

Published by `kyc-service`. Consumed by `wallet-service`.

### KYCSubmitted

Topic: `kyc.submitted.v1` (`topics.KYCSubmitted`)

Published when a customer submits identity documents for verification.

| Field | Type | JSON | Description |
|-------|------|------|-------------|
| `CustomerID` | `string` | `customer_id` | ID of the customer who submitted documents |
| `SubmittedAt` | `time.Time` | `submitted_at` | UTC timestamp of submission |

### KYCVerified

Topic: `kyc.verified.v1` (`topics.KYCVerified`)

Published when an operator approves a KYC verification.
`wallet-service` reacts by activating the customer's account.

| Field | Type | JSON | Description |
|-------|------|------|-------------|
| `CustomerID` | `string` | `customer_id` | ID of the verified customer |
| `VerifiedAt` | `time.Time` | `verified_at` | UTC timestamp of approval |

### KYCRejected

Topic: `kyc.rejected.v1` (`topics.KYCRejected`)

Published when an operator rejects a KYC verification.
`wallet-service` reacts by freezing the customer's account.

| Field | Type | JSON | Description |
|-------|------|------|-------------|
| `CustomerID` | `string` | `customer_id` | ID of the customer whose KYC was rejected |
| `Reason` | `string` | `reason` | Human-readable rejection reason |
| `RejectedAt` | `time.Time` | `rejected_at` | UTC timestamp of rejection |

---

## Wallet Events (`contracts/events/wallet.go`)

Published by `wallet-service`.

### WalletActivated

Topic: `wallet.activated.v1` (`topics.WalletActivated`)

Published when a wallet account transitions to `Active` status (after KYC verified).

| Field | Type | JSON | Description |
|-------|------|------|-------------|
| `AccountID` | `string` | `account_id` | ID of the activated account |
| `CustomerID` | `string` | `customer_id` | ID of the account owner |
| `ActivatedAt` | `time.Time` | `activated_at` | UTC timestamp of activation |

### WalletFrozen

Topic: `wallet.frozen.v1` (`topics.WalletFrozen`)

Published when a wallet account is frozen (after KYC rejected).

| Field | Type | JSON | Description |
|-------|------|------|-------------|
| `AccountID` | `string` | `account_id` | ID of the frozen account |
| `CustomerID` | `string` | `customer_id` | ID of the account owner |
| `Reason` | `string` | `reason` | Reason for freezing (from KYCRejected.Reason) |
| `FrozenAt` | `time.Time` | `frozen_at` | UTC timestamp of freeze |

---

## See Also

- [Contracts Overview](README.md)
- [PLAN-006](../plans/plan-006-event-driven-integration.md) — Kafka wiring for these events
