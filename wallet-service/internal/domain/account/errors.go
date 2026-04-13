package account

import "errors"

var (
	ErrAccountAlreadyExists = errors.New("account: already exists")
	ErrAccountNotFound      = errors.New("account: not found")
	ErrNotActive            = errors.New("account: must be active to perform this operation")
	ErrNotPending           = errors.New("account: must be pending to activate or freeze")
	ErrInsufficientFunds    = errors.New("account: insufficient funds")
)
