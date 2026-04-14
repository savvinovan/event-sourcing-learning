package account

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

// Money is a value object representing an amount in a specific currency.
// Amount and currency are inseparable — there is no money without a currency.
type Money struct {
	Amount   decimal.Decimal
	Currency string // ISO 4217 (e.g. "USD", "EUR", "GBP")
}

var (
	ErrNonPositiveAmount = errors.New("money: amount must be positive")
	ErrCurrencyMismatch  = errors.New("money: currency mismatch")
)

// NewMoney creates a Money value object. Returns error if amount is not positive.
func NewMoney(amount decimal.Decimal, currency string) (Money, error) {
	if amount.LessThanOrEqual(decimal.Zero) {
		return Money{}, ErrNonPositiveAmount
	}
	return Money{Amount: amount, Currency: currency}, nil
}

// Zero returns a zero-amount Money for the given currency.
// Use to initialise an account balance before any transactions.
func Zero(currency string) Money {
	return Money{Amount: decimal.Zero, Currency: currency}
}

// Add returns the sum of two Money values. Currencies must match.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{Amount: m.Amount.Add(other.Amount), Currency: m.Currency}, nil
}

// Sub returns the difference. Currencies must match.
func (m Money) Sub(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{Amount: m.Amount.Sub(other.Amount), Currency: m.Currency}, nil
}

// LessThan reports whether m < other. Currencies must match.
func (m Money) LessThan(other Money) bool {
	return m.Currency == other.Currency && m.Amount.LessThan(other.Amount)
}

// IsPositive reports whether the amount is greater than zero.
func (m Money) IsPositive() bool {
	return m.Amount.GreaterThan(decimal.Zero)
}

// String returns a machine-readable representation: "100.00 USD".
func (m Money) String() string {
	return fmt.Sprintf("%s %s", m.Amount.StringFixed(2), m.Currency)
}

// Display returns a human-readable string formatted per-currency conventions:
// symbol placement (before/after), decimal separator (period/comma), decimal places.
// Falls back to "100.00 CODE" for unknown currencies.
//
//	Money{100, "USD"}.Display() → "$100.00"
//	Money{100, "EUR"}.Display() → "100,00 €"
//	Money{100, "RUB"}.Display() → "100,00 ₽"
//	Money{1000, "JPY"}.Display() → "¥1000"
func (m Money) Display() string {
	f, ok := currencyFormats[m.Currency]
	if !ok {
		// Unknown currency: "100.00 CODE"
		return fmt.Sprintf("%s %s", m.Amount.StringFixed(2), m.Currency)
	}
	formatted := formatAmount(m.Amount, f.decimalSep, f.places)
	if f.symbolAfter {
		return formatted + " " + f.symbol
	}
	return f.symbol + formatted
}

// currencyFmt describes display conventions for a specific currency.
type currencyFmt struct {
	symbol      string
	symbolAfter bool   // true → "100,00 €", false → "$100.00"
	decimalSep  string // "." or ","
	places      int32  // decimal places (e.g. 0 for JPY, 2 for most)
}

// currencyFormats hardcodes display rules for supported currencies.
// Extend this map as new currencies are added.
var currencyFormats = map[string]currencyFmt{
	// Symbol-before, period decimal separator
	"USD": {symbol: "$", symbolAfter: false, decimalSep: ".", places: 2},
	"GBP": {symbol: "£", symbolAfter: false, decimalSep: ".", places: 2},
	"CNY": {symbol: "¥", symbolAfter: false, decimalSep: ".", places: 2},
	"INR": {symbol: "₹", symbolAfter: false, decimalSep: ".", places: 2},
	// No decimal places
	"JPY": {symbol: "¥", symbolAfter: false, decimalSep: ".", places: 0},
	// Symbol-after, comma decimal separator (European convention)
	"EUR": {symbol: "€", symbolAfter: true, decimalSep: ",", places: 2},
	"RUB": {symbol: "₽", symbolAfter: true, decimalSep: ",", places: 2},
	// Code-before, period decimal separator (Swiss convention)
	"CHF": {symbol: "CHF", symbolAfter: false, decimalSep: ".", places: 2},
}

// formatAmount converts a decimal to a string using the given separator and precision.
func formatAmount(amount decimal.Decimal, sep string, places int32) string {
	s := amount.StringFixed(places)
	if sep != "." {
		// Replace the standard period with the locale separator
		for i := len(s) - 1; i >= 0; i-- {
			if s[i] == '.' {
				return s[:i] + sep + s[i+1:]
			}
		}
	}
	return s
}
