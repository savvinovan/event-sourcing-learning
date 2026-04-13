# Domain Layer

## Overview

The domain layer contains the core business logic of the application.
It has **no external dependencies** — only the Go standard library.
All other layers depend on the domain; the domain depends on nothing.

## Bounded Contexts

### Wallet

**Source:** `wallet-service/internal/domain/account/`

The `account` package is the aggregate root for the wallet bounded context.
An `Account` is always rebuilt from its event history — no mutable state is stored in a database.

#### Aggregate: Account

```mermaid
classDiagram
    class Account {
        -customerID CustomerID
        -status AccountStatus
        -balance decimal.Decimal
        -currency string
        +Open(id AccountID, customerID CustomerID, currency)
        +Deposit(amount decimal.Decimal, currency)
        +Withdraw(amount decimal.Decimal, currency)
        +Activate()
        +Freeze(reason)
        +Restore([]DomainEvent)
    }

    class AccountStatus {
        <<enumeration>>
        StatusUnknown
        StatusPending
        StatusActive
        StatusFrozen
    }

    class Money {
        +Amount decimal.Decimal
        +Currency string
        +NewMoney(amount decimal.Decimal, currency)
    }

    Account --> AccountStatus
    Account --> Money
```

#### Domain Events

| Event | Trigger | Fields |
|-------|---------|--------|
| `AccountOpened` | `Open()` | CustomerID `CustomerID`, Currency `string` |
| `MoneyDeposited` | `Deposit()` | Amount `decimal.Decimal`, Currency `string` |
| `MoneyWithdrawn` | `Withdraw()` | Amount `decimal.Decimal`, Currency `string` |
| `AccountActivated` | `Activate()` | — |
| `AccountFrozen` | `Freeze()` | Reason |

#### Status Transitions

```mermaid
stateDiagram-v2
    [*] --> Pending : AccountOpened
    Pending --> Active : AccountActivated (KYC verified)
    Pending --> Frozen : AccountFrozen (KYC rejected)
    Active --> Active : MoneyDeposited / MoneyWithdrawn
```

#### Business Rules

- New accounts start in `Pending` status (awaiting KYC verification).
- Deposits are allowed in any non-frozen status.
- Withdrawals require `Active` status.
- Balance and amounts use `decimal.Decimal` (`github.com/shopspring/decimal`) — never `int64` or `float64`.
- Amount must be positive (`> 0`).
- Currency must match the account's registered currency.
- IDs (`AccountID`, `CustomerID`) are UUID v7 generated via `github.com/google/uuid`.

#### Domain Errors

| Error | Condition |
|-------|-----------|
| `ErrAccountAlreadyExists` | `Open()` called on existing aggregate |
| `ErrAccountNotFound` | No events found for the aggregate ID |
| `ErrNotActive` | Operation requires active (non-frozen) account |
| `ErrNotPending` | `Activate` / `Freeze` require pending status |
| `ErrInsufficientFunds` | Balance < withdrawal amount |
| `ErrCurrencyMismatch` | Deposit/withdraw currency ≠ account currency |
| `ErrNonPositiveAmount` | Amount ≤ 0 |

### Shared Primitives

```mermaid
classDiagram
    class Root {
        -id string
        -version int
        -changes []DomainEvent
        +SetID(id string)
        +ID() string
        +Version() int
        +Changes() []DomainEvent
        +ClearChanges()
        +Record(DomainEvent)
        +LoadFromHistory([]DomainEvent, applyFn)
    }
    class DomainEvent {
        <<interface>>
        +AggregateID() string
        +AggregateType() string
        +EventType() string
        +OccurredAt() time.Time
        +Version() int
    }
    class Base {
        +NewBase(...) Base
    }

    Base ..|> DomainEvent
    Root --> DomainEvent
```

## Contents

- [Aggregate Root](shared/aggregate.md) — `internal/domain/aggregate/aggregate.go`
- [Domain Events](shared/event.md) — `internal/domain/event/event.go`
