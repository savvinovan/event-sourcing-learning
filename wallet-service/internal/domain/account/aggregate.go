package account

import (
	"github.com/savvinovan/wallet-service/internal/domain/aggregate"
	"github.com/savvinovan/wallet-service/internal/domain/event"
)

// Account is the aggregate root for the wallet bounded context.
// State is derived entirely from domain events — never mutated directly.
//
// Design: one Account = one currency.
// An Account is opened in a single ISO 4217 currency and rejects deposits or
// withdrawals in any other currency (ErrCurrencyMismatch).
// Multi-currency wallets are modelled as multiple Accounts (one per currency),
// not as a single Account with a map[currency]balance.
type Account struct {
	aggregate.Root

	customerID CustomerID
	status     AccountStatus
	balance    Money // amount + currency are always kept together
}

// Getters — read-only access for query handlers and projections.
func (a *Account) AccountID() AccountID   { return AccountID(a.ID()) }
func (a *Account) CustomerID() CustomerID { return a.customerID }
func (a *Account) Status() AccountStatus  { return a.status }
func (a *Account) Balance() Money         { return a.balance }
func (a *Account) Currency() string       { return a.balance.Currency }

// Restore rebuilds account state by replaying persisted events.
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
func (a *Account) Deposit(amount Money) error {
	if a.status == StatusFrozen {
		return ErrNotActive
	}
	if !amount.IsPositive() {
		return ErrNonPositiveAmount
	}
	if amount.Currency != a.balance.Currency {
		return ErrCurrencyMismatch
	}
	a.applyAndRecord(MoneyDeposited{
		Base:   event.NewBase(a.ID(), AggregateType, EventTypeMoneyDeposited, a.Version()+1),
		Amount: amount,
	})
	return nil
}

// Withdraw debits funds from the account. Requires Active status and sufficient balance.
func (a *Account) Withdraw(amount Money) error {
	if a.status != StatusActive {
		return ErrNotActive
	}
	if !amount.IsPositive() {
		return ErrNonPositiveAmount
	}
	if amount.Currency != a.balance.Currency {
		return ErrCurrencyMismatch
	}
	if a.balance.LessThan(amount) {
		return ErrInsufficientFunds
	}
	a.applyAndRecord(MoneyWithdrawn{
		Base:   event.NewBase(a.ID(), AggregateType, EventTypeMoneyWithdrawn, a.Version()+1),
		Amount: amount,
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

func (a *Account) applyAndRecord(e event.DomainEvent) {
	a.apply(e)
	a.Record(e)
}

func (a *Account) apply(e event.DomainEvent) {
	switch v := e.(type) {
	case AccountOpened:
		a.SetID(v.AggregateID())
		a.customerID = v.CustomerID
		a.status = StatusPending
		a.balance = Zero(v.Currency)
	case MoneyDeposited:
		a.balance, _ = a.balance.Add(v.Amount) // same currency guaranteed by Deposit()
	case MoneyWithdrawn:
		a.balance, _ = a.balance.Sub(v.Amount) // same currency guaranteed by Withdraw()
	case AccountActivated:
		a.status = StatusActive
	case AccountFrozen:
		a.status = StatusFrozen
	}
}
