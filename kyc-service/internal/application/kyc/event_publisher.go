package kyc

import (
	"context"

	domain "github.com/savvinovan/kyc-service/internal/domain/kyc"
)

// EventPublisher publishes cross-service integration events after KYC decisions.
// Defined in the application layer to keep handlers free of infrastructure imports.
type EventPublisher interface {
	PublishKYCVerified(ctx context.Context, customerID domain.CustomerID) error
	PublishKYCRejected(ctx context.Context, customerID domain.CustomerID, reason string) error
}
