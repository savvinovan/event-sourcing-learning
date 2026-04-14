package kyc

import (
	"context"
	"fmt"

	domain "github.com/savvinovan/kyc-service/internal/domain/kyc"
	"github.com/savvinovan/kyc-service/internal/infrastructure/eventstore"
)

// expectedVersion returns the version before the aggregate's uncommitted changes.
func expectedVersion(agg *domain.KYCVerification) int {
	return agg.Version() - len(agg.Changes())
}

// --- Command Handlers ---

type SubmitKYCHandler struct {
	store eventstore.EventStore
}

type ApproveKYCHandler struct {
	store     eventstore.EventStore
	publisher EventPublisher
}

type RejectKYCHandler struct {
	store     eventstore.EventStore
	publisher EventPublisher
}

func NewSubmitKYCHandler(s eventstore.EventStore) *SubmitKYCHandler {
	return &SubmitKYCHandler{store: s}
}

func NewApproveKYCHandler(s eventstore.EventStore, p EventPublisher) *ApproveKYCHandler {
	return &ApproveKYCHandler{store: s, publisher: p}
}

func NewRejectKYCHandler(s eventstore.EventStore, p EventPublisher) *RejectKYCHandler {
	return &RejectKYCHandler{store: s, publisher: p}
}

func (h *SubmitKYCHandler) Handle(ctx context.Context, cmd SubmitKYCCommand) error {
	events, err := h.store.Load(ctx, string(cmd.VerificationID))
	if err != nil {
		return fmt.Errorf("submit kyc: load: %w", err)
	}
	agg := &domain.KYCVerification{}
	agg.Restore(events)

	if err := agg.Submit(cmd.VerificationID, cmd.CustomerID); err != nil {
		return err
	}
	if err := h.store.Append(ctx, string(cmd.VerificationID), agg.Changes(), expectedVersion(agg)); err != nil {
		return fmt.Errorf("submit kyc: append: %w", err)
	}
	agg.ClearChanges()
	return nil
}

func (h *ApproveKYCHandler) Handle(ctx context.Context, cmd ApproveKYCCommand) error {
	evts, err := h.store.Load(ctx, string(cmd.VerificationID))
	if err != nil {
		return fmt.Errorf("approve kyc: load: %w", err)
	}
	if len(evts) == 0 {
		return domain.ErrVerificationNotFound
	}
	agg := &domain.KYCVerification{}
	agg.Restore(evts)

	if err := agg.Approve(); err != nil {
		return err
	}
	if err := h.store.Append(ctx, string(cmd.VerificationID), agg.Changes(), expectedVersion(agg)); err != nil {
		return fmt.Errorf("approve kyc: append: %w", err)
	}
	agg.ClearChanges()

	// Publish integration event. Domain state is already committed at this point —
	// if Kafka publish fails, the wallet service won't be notified (dual-write hazard).
	// Production fix: use outbox pattern for guaranteed delivery.
	if err := h.publisher.PublishKYCVerified(ctx, agg.CustomerID()); err != nil {
		return fmt.Errorf("approve kyc: publish: %w", err)
	}
	return nil
}

func (h *RejectKYCHandler) Handle(ctx context.Context, cmd RejectKYCCommand) error {
	evts, err := h.store.Load(ctx, string(cmd.VerificationID))
	if err != nil {
		return fmt.Errorf("reject kyc: load: %w", err)
	}
	if len(evts) == 0 {
		return domain.ErrVerificationNotFound
	}
	agg := &domain.KYCVerification{}
	agg.Restore(evts)

	if err := agg.Reject(cmd.Reason); err != nil {
		return err
	}
	if err := h.store.Append(ctx, string(cmd.VerificationID), agg.Changes(), expectedVersion(agg)); err != nil {
		return fmt.Errorf("reject kyc: append: %w", err)
	}
	agg.ClearChanges()

	// See comment in ApproveKYCHandler.Handle about dual-write hazard.
	if err := h.publisher.PublishKYCRejected(ctx, agg.CustomerID(), cmd.Reason); err != nil {
		return fmt.Errorf("reject kyc: publish: %w", err)
	}
	return nil
}
