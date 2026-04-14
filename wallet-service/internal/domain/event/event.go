package event

import "time"

// DomainEvent is the base interface for all domain events.
type DomainEvent interface {
	AggregateID() string
	AggregateType() string
	EventType() string
	OccurredAt() time.Time
	Version() int
}

// Base provides a reusable implementation of DomainEvent metadata.
// Embed it in concrete event structs.
type Base struct {
	aggregateID   string
	aggregateType string
	eventType     string
	occurredAt    time.Time
	version       int
}

func NewBase(aggregateID, aggregateType, eventType string, version int) Base {
	return Base{
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
		eventType:     eventType,
		occurredAt:    time.Now().UTC(),
		version:       version,
	}
}

// RestoreBase reconstructs a Base from persisted storage fields.
// Use this when loading events from the event store — unlike NewBase,
// it preserves the original occurredAt instead of using time.Now().
func RestoreBase(aggregateID, aggregateType, eventType string, version int, occurredAt time.Time) Base {
	return Base{
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
		eventType:     eventType,
		occurredAt:    occurredAt,
		version:       version,
	}
}

func (b Base) AggregateID() string   { return b.aggregateID }
func (b Base) AggregateType() string { return b.aggregateType }
func (b Base) EventType() string     { return b.eventType }
func (b Base) OccurredAt() time.Time { return b.occurredAt }
func (b Base) Version() int          { return b.version }
