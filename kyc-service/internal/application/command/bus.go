package command

import "context"

// CommandType identifies the command for routing through the bus.
type CommandType string

// Command is the marker interface for all commands.
type Command interface {
	CommandType() CommandType
}

// CommandHandler is the type-safe interface for command handler implementations.
type CommandHandler[C Command] interface {
	Handle(ctx context.Context, cmd C) error
}

// Bus dispatches commands to their registered handlers.
type Bus interface {
	Dispatch(ctx context.Context, cmd Command) error
}
