package kyc

import (
	"crypto/rand"
	"fmt"
)

// VerificationID is the unique identifier for a KYCVerification aggregate.
type VerificationID string

// CustomerID is the unique identifier for a customer across bounded contexts.
type CustomerID string

// NewVerificationID generates a new random VerificationID in UUID v4 format.
func NewVerificationID() VerificationID {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return VerificationID(fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]))
}
