package events

import "time"

// WalletActivated is published by wallet-service when an account transitions to Active status.
type WalletActivated struct {
	AccountID   string    `json:"account_id"`
	CustomerID  string    `json:"customer_id"`
	ActivatedAt time.Time `json:"activated_at"`
}

// WalletFrozen is published by wallet-service when an account is frozen due to KYC rejection.
type WalletFrozen struct {
	AccountID  string    `json:"account_id"`
	CustomerID string    `json:"customer_id"`
	Reason     string    `json:"reason"`
	FrozenAt   time.Time `json:"frozen_at"`
}
