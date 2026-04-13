package account

import "errors"

// Money represents a monetary amount in minor units (e.g. cents for USD).
// Amount is always positive; zero is allowed only as an initial balance.
type Money struct {
	Amount   int64  // minor units (e.g. 100 = $1.00 USD)
	Currency string // ISO 4217 currency code (e.g. "USD", "EUR")
}

var (
	ErrNonPositiveAmount = errors.New("money: amount must be positive")
	ErrCurrencyMismatch  = errors.New("money: currency mismatch")
)

func NewMoney(amount int64, currency string) (Money, error) {
	if amount <= 0 {
		return Money{}, ErrNonPositiveAmount
	}
	return Money{Amount: amount, Currency: currency}, nil
}

func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

func (m Money) Sub(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{Amount: m.Amount - other.Amount, Currency: m.Currency}, nil
}
