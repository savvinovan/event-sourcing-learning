package command

import "context"

// Command is the marker interface for all commands.
type Command interface {
	CommandType() string
}

// Handler handles a specific command type.
type Handler interface {
	Handle(ctx context.Context, cmd Command) error
}

// Bus dispatches commands to their registered handlers.
type Bus interface {
	Register(commandType string, handler Handler)
	Dispatch(ctx context.Context, cmd Command) error
}
