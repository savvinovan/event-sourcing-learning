# Project Conventions

## Typed IDs — never use plain `string`

All entity identifiers MUST be typed domain types, not `string`:

```go
type AccountID      string
type CustomerID     string
type VerificationID string
```

This applies in every layer — domain, application, HTTP handler (convert at boundary).

## UUIDs — always UUID v7 via google/uuid

Generate all IDs using `github.com/google/uuid` v7 (time-ordered, good for DB indexing):

```go
import "github.com/google/uuid"

func NewAccountID() AccountID {
    return AccountID(uuid.Must(uuid.NewV7()).String())
}
```

Never use `crypto/rand` + `fmt.Sprintf` manually for ID generation.

## Money — always `decimal.Decimal` from shopspring

Use `github.com/shopspring/decimal` for all monetary amounts. Never use `int64` cents or `float64`:

```go
import "github.com/shopspring/decimal"

type Money struct {
    Amount   decimal.Decimal
    Currency string
}
```

## Code Review

- Run `plannotator review` before every commit (BEFORE staging/committing)
- After launching plannotator, do NOT kill the background task — wait for the user to submit annotations in the chat
- Empty output file = user is still writing, NOT "no feedback"

## Docs as Code

- Every code change must have a matching doc update in `docs/`
- Diagrams in Mermaid, docs in Markdown
