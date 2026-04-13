package eventstore

import (
	"context"

	"github.com/savvinovan/wallet-service/internal/domain/event"
)

// ErrVersionConflict is returned when optimistic concurrency check fails.
var ErrVersionConflict = &versionConflictError{}

type versionConflictError struct{}

func (e *versionConflictError) Error() string { return "event store: version conflict" }

// EventStore persists and loads domain events for an aggregate stream.
type EventStore interface {
	// Append appends events to the aggregate stream.
	// expectedVersion is the version before these events; used for optimistic concurrency.
	Append(ctx context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error

	// Load returns all events for the given aggregate, ordered by version.
	Load(ctx context.Context, aggregateID string) ([]event.DomainEvent, error)
}
