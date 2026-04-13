package kyc

import "github.com/savvinovan/kyc-service/internal/domain/event"

const (
	AggregateType = "kyc_verification"

	EventTypeKYCSubmitted = "KYCSubmitted"
	EventTypeKYCVerified  = "KYCVerified"
	EventTypeKYCRejected  = "KYCRejected"
)

// KYCSubmitted is raised when a customer submits identity documents for verification.
type KYCSubmitted struct {
	event.Base
	CustomerID string
}

// KYCVerified is raised when an operator approves the KYC verification.
type KYCVerified struct {
	event.Base
}

// KYCRejected is raised when an operator rejects the KYC verification.
type KYCRejected struct {
	event.Base
	Reason string
}
