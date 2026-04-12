package aggregate

import "github.com/savvinovan/event-sourcing-learning/internal/domain/event"

// Root is the base for all aggregate roots with event sourcing support.
// Embed it in your aggregate structs and call Record to register domain events.
type Root struct {
	id      string
	version int
	changes []event.DomainEvent
}

func (r *Root) SetID(id string) { r.id = id }
func (r *Root) ID() string      { return r.id }
func (r *Root) Version() int    { return r.version }

// Changes returns uncommitted domain events.
func (r *Root) Changes() []event.DomainEvent { return r.changes }

// ClearChanges discards uncommitted events after they have been persisted.
func (r *Root) ClearChanges() { r.changes = nil }

// Record appends a domain event to the uncommitted changes and bumps the version.
func (r *Root) Record(e event.DomainEvent) {
	r.changes = append(r.changes, e)
	r.version++
}

// LoadFromHistory replays persisted events to restore aggregate state.
// The concrete aggregate must implement Apply.
func (r *Root) LoadFromHistory(events []event.DomainEvent, apply func(event.DomainEvent)) {
	for _, e := range events {
		apply(e)
		r.version = e.Version()
	}
}
