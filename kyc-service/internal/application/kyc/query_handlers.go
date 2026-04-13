package kyc

import (
	"context"
	"fmt"

	domain "github.com/savvinovan/kyc-service/internal/domain/kyc"
	"github.com/savvinovan/kyc-service/internal/infrastructure/eventstore"
)

// GetKYCStatusHandler rebuilds KYCVerification state from the event store.
type GetKYCStatusHandler struct{ store eventstore.EventStore }

func NewGetKYCStatusHandler(s eventstore.EventStore) *GetKYCStatusHandler {
	return &GetKYCStatusHandler{s}
}

func (h *GetKYCStatusHandler) Handle(ctx context.Context, q GetKYCStatusQuery) (KYCStatusResult, error) {
	events, err := h.store.Load(ctx, string(q.VerificationID))
	if err != nil {
		return KYCStatusResult{}, fmt.Errorf("get kyc status: load: %w", err)
	}
	if len(events) == 0 {
		return KYCStatusResult{}, domain.ErrVerificationNotFound
	}
	agg := &domain.KYCVerification{}
	agg.Restore(events)

	return KYCStatusResult{
		VerificationID: agg.VerificationID(),
		CustomerID:     agg.CustomerID(),
		Status:         agg.Status().String(),
		Reason:         agg.Reason(),
	}, nil
}
