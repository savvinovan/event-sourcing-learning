package account

import (
	"context"

	domain "github.com/savvinovan/wallet-service/internal/domain/account"
)

// AccountReadRepository is the read-side port for the account bounded context.
// It queries pre-projected read model tables — no event replay.
// Implemented by infrastructure/readmodel.PostgresReadModelRepository.
type AccountReadRepository interface {
	GetBalance(ctx context.Context, accountID domain.AccountID) (BalanceResult, error)
	GetTransactions(ctx context.Context, accountID domain.AccountID) ([]TransactionRecord, error)
}
