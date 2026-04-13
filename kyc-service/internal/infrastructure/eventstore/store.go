package eventstore

import (
	"context"

	"github.com/savvinovan/kyc-service/internal/domain/event"
)

// ErrVersionConflict is returned when optimistic concurrency check fails.
var ErrVersionConflict = &versionConflictError{}

type versionConflictError struct{}

func (e *versionConflictError) Error() string { return "event store: version conflict" }

// EventStore persists and loads domain events for an aggregate stream.
type EventStore interface {
	Append(ctx context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error
	Load(ctx context.Context, aggregateID string) ([]event.DomainEvent, error)
}
