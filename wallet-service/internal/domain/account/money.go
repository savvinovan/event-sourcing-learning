package account

import (
	"errors"

	"github.com/shopspring/decimal"
)

// Money represents a monetary amount.
// Amount uses decimal.Decimal to avoid floating-point precision issues.
type Money struct {
	Amount   decimal.Decimal
	Currency string // ISO 4217 currency code (e.g. "USD", "EUR")
}

var (
	ErrNonPositiveAmount = errors.New("money: amount must be positive")
	ErrCurrencyMismatch  = errors.New("money: currency mismatch")
)

func NewMoney(amount decimal.Decimal, currency string) (Money, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return Money{}, ErrNonPositiveAmount
	}
	return Money{Amount: amount, Currency: currency}, nil
}

func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{Amount: m.Amount.Add(other.Amount), Currency: m.Currency}, nil
}

func (m Money) Sub(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{Amount: m.Amount.Sub(other.Amount), Currency: m.Currency}, nil
}
