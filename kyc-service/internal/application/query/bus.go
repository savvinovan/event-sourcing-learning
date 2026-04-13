package query

import "context"

// QueryType identifies the query for routing through the bus.
type QueryType string

// Query is the marker interface for all queries.
type Query interface {
	QueryType() QueryType
}

// QueryHandler is the type-safe interface for query handler implementations.
type QueryHandler[Q Query, R any] interface {
	Handle(ctx context.Context, q Q) (R, error)
}

// Bus dispatches queries to their registered handlers.
type Bus interface {
	Ask(ctx context.Context, q Query) (any, error)
}
