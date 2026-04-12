package query

import "context"

// Query is the marker interface for all queries.
type Query interface {
	QueryType() string
}

// Handler handles a specific query type and returns a result.
type Handler interface {
	Handle(ctx context.Context, q Query) (any, error)
}

// Bus dispatches queries to their registered handlers.
type Bus interface {
	Register(queryType string, handler Handler)
	Ask(ctx context.Context, q Query) (any, error)
}
