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

- Every code change MUST have a matching doc update in `docs/` — no exceptions
- Diagrams in Mermaid, docs in Markdown
- **What to update per change type:**
  - New HTTP endpoint → update `docs/interfaces/http.md` or `docs/interfaces/kyc-http.md`
  - New domain event → update `docs/contracts/events.md` and `docs/domain/*.md`
  - New command/query → update `docs/application/commands.md` or `docs/application/queries.md`
  - New infrastructure (DB, queue, etc.) → add/update file in `docs/infrastructure/`
  - New package or OpenAPI spec change → update relevant interface or infrastructure doc
  - Plan completed → mark as completed in `docs/plans/README.md` and the plan file itself
- Run plannotator review after code changes AND after doc updates
