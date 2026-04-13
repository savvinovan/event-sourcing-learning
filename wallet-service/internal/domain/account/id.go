package account

import (
	"crypto/rand"
	"fmt"
)

// AccountID is the unique identifier for an Account aggregate.
type AccountID string

// CustomerID is the unique identifier for a customer across bounded contexts.
type CustomerID string

// NewAccountID generates a new random AccountID in UUID v4 format.
func NewAccountID() AccountID {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return AccountID(fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]))
}
