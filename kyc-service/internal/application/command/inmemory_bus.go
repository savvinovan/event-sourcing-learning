package command

import (
	"context"
	"fmt"
	"sync"
)

type handlerFn func(ctx context.Context, cmd Command) error

// InMemoryCommandBus is a synchronous, in-memory implementation of Bus.
type InMemoryCommandBus struct {
	mu       sync.RWMutex
	handlers map[CommandType]handlerFn
}

func NewInMemoryBus() *InMemoryCommandBus {
	return &InMemoryCommandBus{handlers: make(map[CommandType]handlerFn)}
}

// Dispatch routes a command to its registered handler.
func (b *InMemoryCommandBus) Dispatch(ctx context.Context, cmd Command) error {
	b.mu.RLock()
	h, ok := b.handlers[cmd.CommandType()]
	b.mu.RUnlock()
	if !ok {
		return fmt.Errorf("command bus: no handler registered for %q", cmd.CommandType())
	}
	return h(ctx, cmd)
}

// MustRegister registers a typed CommandHandler[C] with the bus.
// Called at startup; panics if the same command type is registered twice.
func MustRegister[C Command](bus *InMemoryCommandBus, handler CommandHandler[C]) {
	var zero C
	ct := zero.CommandType()

	bus.mu.Lock()
	defer bus.mu.Unlock()

	if _, exists := bus.handlers[ct]; exists {
		panic(fmt.Sprintf("command bus: handler for %q already registered", ct))
	}
	bus.handlers[ct] = func(ctx context.Context, cmd Command) error {
		c, ok := cmd.(C)
		if !ok {
			return fmt.Errorf("command bus: type mismatch for %q: expected %T, got %T", ct, zero, cmd)
		}
		return handler.Handle(ctx, c)
	}
}

// compile-time interface check
var _ Bus = (*InMemoryCommandBus)(nil)
