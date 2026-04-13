package kyc

// KYCStatus represents the current verification status of a KYCVerification aggregate.
type KYCStatus int

const (
	StatusUnknown   KYCStatus = iota // 0 — zero value, invalid/unset
	StatusSubmitted                  // 1 — awaiting operator decision
	StatusVerified                   // 2 — approved by operator
	StatusRejected                   // 3 — rejected by operator
)

func (s KYCStatus) String() string {
	switch s {
	case StatusSubmitted:
		return "Submitted"
	case StatusVerified:
		return "Verified"
	case StatusRejected:
		return "Rejected"
	default:
		return "Unknown"
	}
}
