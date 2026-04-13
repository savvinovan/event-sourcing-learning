package kyc

import (
	"github.com/savvinovan/kyc-service/internal/domain/aggregate"
	"github.com/savvinovan/kyc-service/internal/domain/event"
)

// KYCVerification is the aggregate root for the KYC bounded context.
// State is derived entirely from domain events — never mutated directly.
type KYCVerification struct {
	aggregate.Root

	customerID string
	status     KYCStatus
	reason     string // set on rejection
}

// Getters — read-only access for query handlers and projections.
func (k *KYCVerification) CustomerID() string { return k.customerID }
func (k *KYCVerification) Status() KYCStatus  { return k.status }
func (k *KYCVerification) Reason() string     { return k.reason }

// Restore rebuilds verification state by replaying persisted events.
func (k *KYCVerification) Restore(events []event.DomainEvent) {
	k.Root.LoadFromHistory(events, k.apply)
}

// --- Commands ---

// Submit creates a new KYC verification for a customer.
func (k *KYCVerification) Submit(id, customerID string) error {
	if k.ID() != "" {
		return ErrVerificationAlreadyExists
	}
	k.applyAndRecord(KYCSubmitted{
		Base:       event.NewBase(id, AggregateType, EventTypeKYCSubmitted, k.Version()+1),
		CustomerID: customerID,
	})
	return nil
}

// Approve transitions the verification from Submitted to Verified.
func (k *KYCVerification) Approve() error {
	if k.status == StatusVerified {
		return ErrAlreadyVerified
	}
	if k.status != StatusSubmitted {
		return ErrNotSubmitted
	}
	k.applyAndRecord(KYCVerified{
		Base: event.NewBase(k.ID(), AggregateType, EventTypeKYCVerified, k.Version()+1),
	})
	return nil
}

// Reject transitions the verification from Submitted to Rejected.
func (k *KYCVerification) Reject(reason string) error {
	if k.status == StatusRejected {
		return ErrAlreadyRejected
	}
	if k.status != StatusSubmitted {
		return ErrNotSubmitted
	}
	k.applyAndRecord(KYCRejected{
		Base:   event.NewBase(k.ID(), AggregateType, EventTypeKYCRejected, k.Version()+1),
		Reason: reason,
	})
	return nil
}

// --- Event sourcing internals ---

func (k *KYCVerification) applyAndRecord(e event.DomainEvent) {
	k.apply(e)
	k.Record(e)
}

func (k *KYCVerification) apply(e event.DomainEvent) {
	switch v := e.(type) {
	case KYCSubmitted:
		k.SetID(v.AggregateID())
		k.customerID = v.CustomerID
		k.status = StatusSubmitted
	case KYCVerified:
		k.status = StatusVerified
	case KYCRejected:
		k.status = StatusRejected
		k.reason = v.Reason
	}
}
