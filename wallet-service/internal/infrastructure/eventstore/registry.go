package eventstore

import (
	"fmt"

	"github.com/savvinovan/wallet-service/internal/domain/event"
)

// SerializeFunc encodes a domain event into its JSONB payload bytes.
type SerializeFunc func(event.DomainEvent) ([]byte, error)

// DeserializeFunc reconstructs a domain event from persisted metadata and payload bytes.
type DeserializeFunc func(base event.Base, payload []byte) (event.DomainEvent, error)

// eventCodec pairs serializer and deserializer for one event type.
type eventCodec struct {
	version     int
	serialize   SerializeFunc
	deserialize DeserializeFunc
}

// Registry maps event type strings to their codecs.
// Each service builds one registry at startup and injects it into the event store.
type Registry struct {
	codecs map[string]eventCodec
}

func NewRegistry() *Registry {
	return &Registry{codecs: make(map[string]eventCodec)}
}

// Register associates an event type string with its serializer and deserializer at schema version 1.
func (r *Registry) Register(eventType string, s SerializeFunc, d DeserializeFunc) {
	r.RegisterV(eventType, 1, s, d)
}

// RegisterV is like Register but lets the caller declare the current schema version.
// Use when introducing a new payload shape (v2, v3, …) so that Append writes the
// correct schema_version and the UpcasterRegistry can chain transforms for old rows.
func (r *Registry) RegisterV(eventType string, version int, s SerializeFunc, d DeserializeFunc) {
	r.codecs[eventType] = eventCodec{version: version, serialize: s, deserialize: d}
}

// GetLatestVersion returns the schema version the codec expects for eventType.
// Returns 1 if no codec is registered (safe default for unknown types).
func (r *Registry) GetLatestVersion(eventType string) int {
	c, ok := r.codecs[eventType]
	if !ok {
		return 1
	}
	return c.version
}

// Serialize encodes the event into JSONB payload bytes.
func (r *Registry) Serialize(e event.DomainEvent) ([]byte, error) {
	c, ok := r.codecs[e.EventType()]
	if !ok {
		return nil, fmt.Errorf("event registry: no codec registered for %q", e.EventType())
	}
	return c.serialize(e)
}

// Deserialize reconstructs a domain event from its type, base metadata, and payload bytes.
// Returns an error (not a panic) for unknown event types.
func (r *Registry) Deserialize(eventType string, base event.Base, payload []byte) (event.DomainEvent, error) {
	c, ok := r.codecs[eventType]
	if !ok {
		return nil, fmt.Errorf("event registry: unknown event type %q", eventType)
	}
	return c.deserialize(base, payload)
}
