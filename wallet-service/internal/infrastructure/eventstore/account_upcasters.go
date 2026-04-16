package eventstore

import (
	"encoding/json"
	"fmt"

	domain "github.com/savvinovan/wallet-service/internal/domain/account"
)

// NewAccountUpcasterRegistry builds upcasters for all wallet domain event schema migrations.
func NewAccountUpcasterRegistry() *UpcasterRegistry {
	r := NewUpcasterRegistry()
	r.Register(domain.EventTypeMoneyDeposited, 1, upcastMoneyDepositedV1ToV2)
	return r
}

// upcastMoneyDepositedV1ToV2 adds the "description" field (default "") to old
// MoneyDeposited payloads written before schema v2.
//
// v1 shape: {"amount":"…","currency":"…"}
// v2 shape: {"amount":"…","currency":"…","description":""}
func upcastMoneyDepositedV1ToV2(payload []byte) ([]byte, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal MoneyDeposited v1 payload: %w", err)
	}
	if _, ok := raw["description"]; !ok {
		raw["description"] = json.RawMessage(`""`)
	}
	return json.Marshal(raw)
}
