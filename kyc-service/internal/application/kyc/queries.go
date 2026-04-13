package kyc

import "github.com/savvinovan/kyc-service/internal/application/query"

const (
	QryGetKYCStatus query.QueryType = "GetKYCStatus"
)

// GetKYCStatusQuery returns the current status of a KYC verification.
type GetKYCStatusQuery struct {
	VerificationID string
}

func (q GetKYCStatusQuery) QueryType() query.QueryType { return QryGetKYCStatus }

// KYCStatusResult is the read model returned by GetKYCStatusQuery.
type KYCStatusResult struct {
	VerificationID string
	CustomerID     string
	Status         string
	Reason         string // non-empty only when rejected
}
