package account

import "github.com/savvinovan/wallet-service/internal/domain/event"

const AggregateType = "Account"

const (
	EventTypeAccountOpened    = "AccountOpened"
	EventTypeMoneyDeposited   = "MoneyDeposited"
	EventTypeMoneyWithdrawn   = "MoneyWithdrawn"
	EventTypeAccountActivated = "AccountActivated"
	EventTypeAccountFrozen    = "AccountFrozen"
)

// AccountOpened is recorded when a new account is created for a customer.
type AccountOpened struct {
	event.Base
	CustomerID CustomerID
	Currency   string
}

// MoneyDeposited is recorded when funds are credited to an account.
type MoneyDeposited struct {
	event.Base
	Amount      Money
	Description string // schema v2: optional human-readable label; defaults to "" for v1 rows
}

// MoneyWithdrawn is recorded when funds are debited from an account.
type MoneyWithdrawn struct {
	event.Base
	Amount Money
}

// AccountActivated is recorded when KYC verification is approved.
type AccountActivated struct {
	event.Base
}

// AccountFrozen is recorded when KYC verification is rejected or the account is frozen manually.
type AccountFrozen struct {
	event.Base
	Reason string
}
