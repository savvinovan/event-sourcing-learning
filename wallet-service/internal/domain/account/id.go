package account

import "github.com/google/uuid"

// AccountID is the unique identifier for an Account aggregate.
type AccountID string

// CustomerID is the unique identifier for a customer across bounded contexts.
type CustomerID string

// NewAccountID generates a new time-ordered AccountID using UUID v7.
func NewAccountID() AccountID {
	return AccountID(uuid.Must(uuid.NewV7()).String())
}
