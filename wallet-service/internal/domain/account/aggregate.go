package account

import (
	"github.com/savvinovan/wallet-service/internal/domain/aggregate"
	"github.com/savvinovan/wallet-service/internal/domain/event"
)

// Account is the aggregate root for the wallet bounded context.
// State is derived entirely from domain events — never mutated directly.
type Account struct {
	aggregate.Root

	customerID CustomerID
	status     AccountStatus
	balance    int64  // in minor units (e.g. cents)
	currency   string
}

// Getters — read-only access for query handlers and projections.
func (a *Account) AccountID() AccountID     { return AccountID(a.ID()) }
func (a *Account) CustomerID() CustomerID   { return a.customerID }
func (a *Account) Status() AccountStatus    { return a.status }
func (a *Account) Balance() int64           { return a.balance }
func (a *Account) Currency() string         { return a.currency }

// Restore rebuilds account state by replaying persisted events.
// Called by command and query handlers before executing any operation.
func (a *Account) Restore(events []event.DomainEvent) {
	a.Root.LoadFromHistory(events, a.apply)
}

// --- Commands ---

// Open creates a new account for a customer in Pending status (awaiting KYC).
func (a *Account) Open(id AccountID, customerID CustomerID, currency string) error {
	if a.ID() != "" {
		return ErrAccountAlreadyExists
	}
	a.applyAndRecord(AccountOpened{
		Base:       event.NewBase(string(id), AggregateType, EventTypeAccountOpened, a.Version()+1),
		CustomerID: customerID,
		Currency:   currency,
	})
	return nil
}

// Deposit credits funds to the account. Allowed in any non-frozen status.
func (a *Account) Deposit(amount int64, currency string) error {
	if a.status == StatusFrozen {
		return ErrNotActive
	}
	if _, err := NewMoney(amount, currency); err != nil {
		return err
	}
	if currency != a.currency {
		return ErrCurrencyMismatch
	}
	a.applyAndRecord(MoneyDeposited{
		Base:     event.NewBase(a.ID(), AggregateType, EventTypeMoneyDeposited, a.Version()+1),
		Amount:   amount,
		Currency: currency,
	})
	return nil
}

// Withdraw debits funds from the account. Requires Active status and sufficient balance.
func (a *Account) Withdraw(amount int64, currency string) error {
	if a.status != StatusActive {
		return ErrNotActive
	}
	if _, err := NewMoney(amount, currency); err != nil {
		return err
	}
	if currency != a.currency {
		return ErrCurrencyMismatch
	}
	if a.balance < amount {
		return ErrInsufficientFunds
	}
	a.applyAndRecord(MoneyWithdrawn{
		Base:     event.NewBase(a.ID(), AggregateType, EventTypeMoneyWithdrawn, a.Version()+1),
		Amount:   amount,
		Currency: currency,
	})
	return nil
}

// Activate transitions the account from Pending to Active (triggered by KYCVerified).
func (a *Account) Activate() error {
	if a.status != StatusPending {
		return ErrNotPending
	}
	a.applyAndRecord(AccountActivated{
		Base: event.NewBase(a.ID(), AggregateType, EventTypeAccountActivated, a.Version()+1),
	})
	return nil
}

// Freeze transitions the account to Frozen (triggered by KYCRejected).
func (a *Account) Freeze(reason string) error {
	if a.status != StatusPending {
		return ErrNotPending
	}
	a.applyAndRecord(AccountFrozen{
		Base:   event.NewBase(a.ID(), AggregateType, EventTypeAccountFrozen, a.Version()+1),
		Reason: reason,
	})
	return nil
}

// --- Event sourcing internals ---

// applyAndRecord applies a domain event to update state and records it as an uncommitted change.
func (a *Account) applyAndRecord(e event.DomainEvent) {
	a.apply(e)
	a.Record(e)
}

// apply updates aggregate state from a domain event.
// Called both during command execution and history replay.
func (a *Account) apply(e event.DomainEvent) {
	switch v := e.(type) {
	case AccountOpened:
		a.SetID(v.AggregateID())
		a.customerID = v.CustomerID
		a.currency = v.Currency
		a.status = StatusPending
		a.balance = 0
	case MoneyDeposited:
		a.balance += v.Amount
	case MoneyWithdrawn:
		a.balance -= v.Amount
	case AccountActivated:
		a.status = StatusActive
	case AccountFrozen:
		a.status = StatusFrozen
	}
}
