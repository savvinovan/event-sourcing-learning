package command

import "context"

// CommandType identifies the command for routing through the bus.
type CommandType string

// Command is the marker interface for all commands.
type Command interface {
	CommandType() CommandType
}

// CommandHandler is the type-safe interface for command handler implementations.
// C is the concrete command type.
// Implementations receive strongly-typed arguments — no type assertions needed inside the handler.
type CommandHandler[C Command] interface {
	Handle(ctx context.Context, cmd C) error
}

// Bus dispatches commands to registered handlers.
// Use the package-level MustRegister function to register handlers at startup.
type Bus interface {
	Dispatch(ctx context.Context, cmd Command) error
}
