// Package messaging contains Kafka publisher stubs for cross-service events.
// Full implementation: see PLAN-006 (event-driven integration).
package messaging

import (
	"github.com/savvinovan/event-sourcing-learning/contracts/events"
	"github.com/savvinovan/event-sourcing-learning/contracts/topics"
)

// publishKYCVerified publishes a KYCVerified event to Kafka.
// Full wiring: PLAN-006.
func publishKYCVerified(e events.KYCVerified) (topic string) {
	// Fields used explicitly — compile error if contracts schema changes:
	_ = e.CustomerID
	_ = e.VerifiedAt
	return topics.KYCVerified
}

// publishKYCRejected publishes a KYCRejected event to Kafka.
// Full wiring: PLAN-006.
func publishKYCRejected(e events.KYCRejected) (topic string) {
	// Fields used explicitly — compile error if contracts schema changes:
	_ = e.CustomerID
	_ = e.Reason
	_ = e.RejectedAt
	return topics.KYCRejected
}

// publishKYCSubmitted publishes a KYCSubmitted event to Kafka.
// Full wiring: PLAN-006.
func publishKYCSubmitted(e events.KYCSubmitted) (topic string) {
	// Fields used explicitly — compile error if contracts schema changes:
	_ = e.CustomerID
	_ = e.SubmittedAt
	return topics.KYCSubmitted
}
