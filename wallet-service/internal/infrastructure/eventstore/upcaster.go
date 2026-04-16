package eventstore

import "fmt"

// UpcastFunc transforms a payload at schema version N into version N+1.
type UpcastFunc func(payload []byte) ([]byte, error)

type upcasterKey struct {
	eventType     string
	schemaVersion int
}

// UpcasterRegistry holds upcasters keyed by (eventType, fromSchemaVersion).
// When loading old events from the DB, Upcast chains transforms until the
// payload reaches the version the Registry codec expects (latest).
type UpcasterRegistry struct {
	upcasters map[upcasterKey]UpcastFunc
}

func NewUpcasterRegistry() *UpcasterRegistry {
	return &UpcasterRegistry{upcasters: make(map[upcasterKey]UpcastFunc)}
}

// Register adds an upcaster from fromVersion → fromVersion+1 for eventType.
func (r *UpcasterRegistry) Register(eventType string, fromVersion int, fn UpcastFunc) {
	r.upcasters[upcasterKey{eventType: eventType, schemaVersion: fromVersion}] = fn
}

// Upcast chains upcasters starting at schemaVersion until no further upcaster
// is registered for (eventType, currentVersion). Returns the final payload
// ready for Registry.Deserialize, which always expects the latest shape.
func (r *UpcasterRegistry) Upcast(eventType string, schemaVersion int, payload []byte) ([]byte, error) {
	p := payload
	for current := schemaVersion; ; current++ {
		fn, ok := r.upcasters[upcasterKey{eventType: eventType, schemaVersion: current}]
		if !ok {
			return p, nil
		}
		var err error
		p, err = fn(p)
		if err != nil {
			return nil, fmt.Errorf("upcast %s v%d→v%d: %w", eventType, current, current+1, err)
		}
	}
}
