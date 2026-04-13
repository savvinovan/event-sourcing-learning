package eventstore

import (
	"context"
	"sync"

	"github.com/savvinovan/wallet-service/internal/domain/event"
)

// InMemoryEventStore is a thread-safe in-memory implementation of EventStore.
// Suitable for development and testing; not persistent across restarts.
type InMemoryEventStore struct {
	mu      sync.RWMutex
	streams map[string][]event.DomainEvent
}

func NewInMemory() *InMemoryEventStore {
	return &InMemoryEventStore{
		streams: make(map[string][]event.DomainEvent),
	}
}

// Append appends events to the aggregate stream with optimistic concurrency control.
// Returns ErrVersionConflict if the current stream version != expectedVersion.
func (s *InMemoryEventStore) Append(_ context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing := s.streams[aggregateID]
	if len(existing) != expectedVersion {
		return ErrVersionConflict
	}

	s.streams[aggregateID] = append(existing, events...)
	return nil
}

// Load returns all events for the given aggregate in version order.
func (s *InMemoryEventStore) Load(_ context.Context, aggregateID string) ([]event.DomainEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	events := s.streams[aggregateID]
	if len(events) == 0 {
		return nil, nil
	}
	// return a copy to prevent external mutation
	result := make([]event.DomainEvent, len(events))
	copy(result, events)
	return result, nil
}

// compile-time interface check
var _ EventStore = (*InMemoryEventStore)(nil)
