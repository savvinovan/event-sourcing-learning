package eventstore

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"

	domain "github.com/savvinovan/wallet-service/internal/domain/account"
	"github.com/savvinovan/wallet-service/internal/domain/event"
)

// NewAccountRegistry builds a Registry with codecs for all wallet domain event types.
func NewAccountRegistry() *Registry {
	r := NewRegistry()

	r.Register(domain.EventTypeAccountOpened,
		func(e event.DomainEvent) ([]byte, error) {
			v := e.(domain.AccountOpened)
			return json.Marshal(struct {
				CustomerID string `json:"customer_id"`
				Currency   string `json:"currency"`
			}{
				CustomerID: string(v.CustomerID),
				Currency:   v.Currency,
			})
		},
		func(base event.Base, payload []byte) (event.DomainEvent, error) {
			var p struct {
				CustomerID string `json:"customer_id"`
				Currency   string `json:"currency"`
			}
			if err := json.Unmarshal(payload, &p); err != nil {
				return nil, fmt.Errorf("deserialize AccountOpened: %w", err)
			}
			return domain.AccountOpened{
				Base:       base,
				CustomerID: domain.CustomerID(p.CustomerID),
				Currency:   p.Currency,
			}, nil
		},
	)

	r.Register(domain.EventTypeMoneyDeposited,
		func(e event.DomainEvent) ([]byte, error) {
			return serializeMoney(e.(domain.MoneyDeposited).Amount)
		},
		func(base event.Base, payload []byte) (event.DomainEvent, error) {
			m, err := deserializeMoney(payload)
			if err != nil {
				return nil, fmt.Errorf("deserialize MoneyDeposited: %w", err)
			}
			return domain.MoneyDeposited{Base: base, Amount: m}, nil
		},
	)

	r.Register(domain.EventTypeMoneyWithdrawn,
		func(e event.DomainEvent) ([]byte, error) {
			return serializeMoney(e.(domain.MoneyWithdrawn).Amount)
		},
		func(base event.Base, payload []byte) (event.DomainEvent, error) {
			m, err := deserializeMoney(payload)
			if err != nil {
				return nil, fmt.Errorf("deserialize MoneyWithdrawn: %w", err)
			}
			return domain.MoneyWithdrawn{Base: base, Amount: m}, nil
		},
	)

	r.Register(domain.EventTypeAccountActivated,
		func(e event.DomainEvent) ([]byte, error) { return []byte("{}"), nil },
		func(base event.Base, _ []byte) (event.DomainEvent, error) {
			return domain.AccountActivated{Base: base}, nil
		},
	)

	r.Register(domain.EventTypeAccountFrozen,
		func(e event.DomainEvent) ([]byte, error) {
			v := e.(domain.AccountFrozen)
			return json.Marshal(struct {
				Reason string `json:"reason"`
			}{Reason: v.Reason})
		},
		func(base event.Base, payload []byte) (event.DomainEvent, error) {
			var p struct {
				Reason string `json:"reason"`
			}
			if err := json.Unmarshal(payload, &p); err != nil {
				return nil, fmt.Errorf("deserialize AccountFrozen: %w", err)
			}
			return domain.AccountFrozen{Base: base, Reason: p.Reason}, nil
		},
	)

	return r
}

// moneyPayload is the JSONB representation of a Money value object.
// Amount is stored as a string to preserve decimal precision (never as a JSON number).
type moneyPayload struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

func serializeMoney(m domain.Money) ([]byte, error) {
	return json.Marshal(moneyPayload{
		Amount:   m.Amount.String(),
		Currency: m.Currency,
	})
}

func deserializeMoney(payload []byte) (domain.Money, error) {
	var p moneyPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return domain.Money{}, err
	}
	amount, err := decimal.NewFromString(p.Amount)
	if err != nil {
		return domain.Money{}, fmt.Errorf("parse decimal %q: %w", p.Amount, err)
	}
	return domain.Money{Amount: amount, Currency: p.Currency}, nil
}
