package account

import "github.com/savvinovan/wallet-service/internal/domain/event"

const AggregateType = "Account"

// Domain event type string constants — used for deserialization and type switches.
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
	Amount   int64
	Currency string
}

// MoneyWithdrawn is recorded when funds are debited from an account.
type MoneyWithdrawn struct {
	event.Base
	Amount   int64
	Currency string
}

// AccountActivated is recorded when KYC verification is approved.
// After this event the account accepts withdrawals and transfers.
type AccountActivated struct {
	event.Base
}

// AccountFrozen is recorded when KYC verification is rejected or the account is frozen manually.
// After this event the account rejects all mutations.
type AccountFrozen struct {
	event.Base
	Reason string
}
