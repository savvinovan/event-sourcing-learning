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

// Bus dispatches commands to their registered handlers.
// Register accepts any CommandHandler[C] implementation.
type Bus interface {
	Register(commandType CommandType, handler any)
	Dispatch(ctx context.Context, cmd Command) error
}
