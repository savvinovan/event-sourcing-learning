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

type SubmitKYCHandler struct{ store eventstore.EventStore }
type ApproveKYCHandler struct{ store eventstore.EventStore }
type RejectKYCHandler struct{ store eventstore.EventStore }

func NewSubmitKYCHandler(s eventstore.EventStore) *SubmitKYCHandler   { return &SubmitKYCHandler{s} }
func NewApproveKYCHandler(s eventstore.EventStore) *ApproveKYCHandler { return &ApproveKYCHandler{s} }
func NewRejectKYCHandler(s eventstore.EventStore) *RejectKYCHandler   { return &RejectKYCHandler{s} }

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
	events, err := h.store.Load(ctx, string(cmd.VerificationID))
	if err != nil {
		return fmt.Errorf("approve kyc: load: %w", err)
	}
	if len(events) == 0 {
		return domain.ErrVerificationNotFound
	}
	agg := &domain.KYCVerification{}
	agg.Restore(events)

	if err := agg.Approve(); err != nil {
		return err
	}
	if err := h.store.Append(ctx, string(cmd.VerificationID), agg.Changes(), expectedVersion(agg)); err != nil {
		return fmt.Errorf("approve kyc: append: %w", err)
	}
	agg.ClearChanges()
	return nil
}

func (h *RejectKYCHandler) Handle(ctx context.Context, cmd RejectKYCCommand) error {
	events, err := h.store.Load(ctx, string(cmd.VerificationID))
	if err != nil {
		return fmt.Errorf("reject kyc: load: %w", err)
	}
	if len(events) == 0 {
		return domain.ErrVerificationNotFound
	}
	agg := &domain.KYCVerification{}
	agg.Restore(events)

	if err := agg.Reject(cmd.Reason); err != nil {
		return err
	}
	if err := h.store.Append(ctx, string(cmd.VerificationID), agg.Changes(), expectedVersion(agg)); err != nil {
		return fmt.Errorf("reject kyc: append: %w", err)
	}
	agg.ClearChanges()
	return nil
}
