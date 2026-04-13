package query

import (
	"context"
	"fmt"
	"sync"
)

type handlerFn func(ctx context.Context, q Query) (any, error)

// InMemoryQueryBus is a synchronous, in-memory implementation of Bus.
type InMemoryQueryBus struct {
	mu       sync.RWMutex
	handlers map[QueryType]handlerFn
}

func NewInMemoryBus() *InMemoryQueryBus {
	return &InMemoryQueryBus{handlers: make(map[QueryType]handlerFn)}
}

// Ask routes a query to its registered handler and returns the result.
func (b *InMemoryQueryBus) Ask(ctx context.Context, q Query) (any, error) {
	b.mu.RLock()
	h, ok := b.handlers[q.QueryType()]
	b.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("query bus: no handler registered for %q", q.QueryType())
	}
	return h(ctx, q)
}

// MustRegister registers a typed QueryHandler[Q, R] with the bus.
// Called at startup; panics if the same query type is registered twice.
func MustRegister[Q Query, R any](bus *InMemoryQueryBus, handler QueryHandler[Q, R]) {
	var zero Q
	qt := zero.QueryType()

	bus.mu.Lock()
	defer bus.mu.Unlock()

	if _, exists := bus.handlers[qt]; exists {
		panic(fmt.Sprintf("query bus: handler for %q already registered", qt))
	}
	bus.handlers[qt] = func(ctx context.Context, q Query) (any, error) {
		typed, ok := q.(Q)
		if !ok {
			return nil, fmt.Errorf("query bus: type mismatch for %q: expected %T, got %T", qt, zero, q)
		}
		return handler.Handle(ctx, typed)
	}
}

// compile-time interface check
var _ Bus = (*InMemoryQueryBus)(nil)
