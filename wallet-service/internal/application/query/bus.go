package query

import "context"

// QueryType identifies the query for routing through the bus.
type QueryType string

// Query is the marker interface for all queries.
type Query interface {
	QueryType() QueryType
}

// QueryHandler is the type-safe interface for query handler implementations.
// Q is the concrete query type, R is the result type.
// Implementations are registered with Bus and receive strongly-typed arguments.
type QueryHandler[Q Query, R any] interface {
	Handle(ctx context.Context, q Q) (R, error)
}

// Bus dispatches queries to registered handlers.
// Use the package-level MustRegister function to register handlers at startup.
// Ask returns any; callers type-assert to the expected result type.
type Bus interface {
	Ask(ctx context.Context, q Query) (any, error)
}
