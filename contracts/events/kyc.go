package events

import "time"

// KYCSubmitted is published by kyc-service when a customer submits identity documents.
type KYCSubmitted struct {
	CustomerID  string    `json:"customer_id"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// KYCVerified is published by kyc-service when an operator approves KYC verification.
// wallet-service subscribes to this event to activate the customer's account.
type KYCVerified struct {
	CustomerID string    `json:"customer_id"`
	VerifiedAt time.Time `json:"verified_at"`
}

// KYCRejected is published by kyc-service when an operator rejects KYC verification.
// wallet-service subscribes to this event to freeze the customer's account.
type KYCRejected struct {
	CustomerID string    `json:"customer_id"`
	Reason     string    `json:"reason"`
	RejectedAt time.Time `json:"rejected_at"`
}
