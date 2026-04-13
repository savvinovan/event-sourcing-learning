package kyc

import (
	"github.com/savvinovan/kyc-service/internal/application/query"
	domain "github.com/savvinovan/kyc-service/internal/domain/kyc"
)

const (
	QryGetKYCStatus query.QueryType = "GetKYCStatus"
)

// GetKYCStatusQuery returns the current status of a KYC verification.
type GetKYCStatusQuery struct {
	VerificationID domain.VerificationID
}

func (q GetKYCStatusQuery) QueryType() query.QueryType { return QryGetKYCStatus }

// KYCStatusResult is the read model returned by GetKYCStatusQuery.
type KYCStatusResult struct {
	VerificationID domain.VerificationID
	CustomerID     domain.CustomerID
	Status         string
	Reason         string // non-empty only when rejected
}
