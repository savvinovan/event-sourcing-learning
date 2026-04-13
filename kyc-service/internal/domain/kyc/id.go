package kyc

import "github.com/google/uuid"

// VerificationID is the unique identifier for a KYCVerification aggregate.
type VerificationID string

// CustomerID is the unique identifier for a customer across bounded contexts.
type CustomerID string

// NewVerificationID generates a new time-ordered VerificationID using UUID v7.
func NewVerificationID() VerificationID {
	return VerificationID(uuid.Must(uuid.NewV7()).String())
}
