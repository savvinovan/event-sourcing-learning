package aggregate

import "github.com/savvinovan/kyc-service/internal/domain/event"

// Root is the base for all aggregate roots with event sourcing support.
type Root struct {
	id      string
	version int
	changes []event.DomainEvent
}

func (r *Root) SetID(id string) { r.id = id }
func (r *Root) ID() string      { return r.id }
func (r *Root) Version() int    { return r.version }

func (r *Root) Changes() []event.DomainEvent { return r.changes }
func (r *Root) ClearChanges()                { r.changes = nil }

func (r *Root) Record(e event.DomainEvent) {
	r.changes = append(r.changes, e)
	r.version++
}

func (r *Root) LoadFromHistory(events []event.DomainEvent, apply func(event.DomainEvent)) {
	for _, e := range events {
		apply(e)
		r.version = e.Version()
	}
}
