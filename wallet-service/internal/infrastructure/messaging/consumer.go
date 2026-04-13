// Package messaging contains Kafka consumer stubs for cross-service events.
// Full implementation: see PLAN-006 (event-driven integration).
package messaging

import (
	"github.com/savvinovan/event-sourcing-learning/contracts/events"
	"github.com/savvinovan/event-sourcing-learning/contracts/topics"
)

// KYCEventTopics lists Kafka topics that wallet-service subscribes to.
var KYCEventTopics = []string{
	topics.KYCVerified,
	topics.KYCRejected,
}

// onKYCVerified is called when a KYCVerified event is consumed from Kafka.
// It maps the contract event to an ActivateAccount command.
// Full wiring: PLAN-006.
func onKYCVerified(e events.KYCVerified) {
	// Fields used explicitly — compile error if contracts schema changes:
	_ = e.CustomerID
	_ = e.VerifiedAt
}

// onKYCRejected is called when a KYCRejected event is consumed from Kafka.
// It maps the contract event to a FreezeAccount command.
// Full wiring: PLAN-006.
func onKYCRejected(e events.KYCRejected) {
	// Fields used explicitly — compile error if contracts schema changes:
	_ = e.CustomerID
	_ = e.Reason
	_ = e.RejectedAt
}
