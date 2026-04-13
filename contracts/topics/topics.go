package topics

// Kafka topic names for cross-service event communication.
// Topics are versioned (e.g. "kyc.submitted.v1") to allow introducing a new schema
// version on a separate topic without breaking existing consumers.
// Both producer and consumer must use the same constant — changing a topic name here
// is a breaking change that requires updating both services simultaneously.
const (
	KYCSubmitted = "kyc.submitted.v1"
	KYCVerified  = "kyc.verified.v1"
	KYCRejected  = "kyc.rejected.v1"

	WalletActivated = "wallet.activated.v1"
	WalletFrozen    = "wallet.frozen.v1"
)
