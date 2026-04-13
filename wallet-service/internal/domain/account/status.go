package account

// AccountStatus represents the lifecycle state of a wallet account.
type AccountStatus int

const (
	StatusUnknown AccountStatus = iota // 0 — invalid/zero value, never set explicitly
	StatusPending                       // 1 — account opened, awaiting KYC verification
	StatusActive                        // 2 — KYC verified, transactions allowed
	StatusFrozen                        // 3 — KYC rejected or manually frozen
)

func (s AccountStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusActive:
		return "active"
	case StatusFrozen:
		return "frozen"
	default:
		return "unknown"
	}
}
