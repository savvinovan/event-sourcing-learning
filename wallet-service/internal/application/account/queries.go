package account

import (
	"github.com/savvinovan/wallet-service/internal/application/query"
	domain "github.com/savvinovan/wallet-service/internal/domain/account"
)

const (
	QryGetBalance      query.QueryType = "GetBalance"
	QryGetTransactions query.QueryType = "GetTransactions"
)

// GetBalanceQuery returns the current balance and status of an account.
type GetBalanceQuery struct {
	AccountID domain.AccountID
}

func (q GetBalanceQuery) QueryType() query.QueryType { return QryGetBalance }

// BalanceResult is the read model returned by GetBalanceQuery.
type BalanceResult struct {
	AccountID  domain.AccountID
	CustomerID domain.CustomerID
	Balance    int64
	Currency   string
	Status     string
}

// GetTransactionsQuery returns the transaction history for an account.
type GetTransactionsQuery struct {
	AccountID domain.AccountID
}

func (q GetTransactionsQuery) QueryType() query.QueryType { return QryGetTransactions }

// TransactionRecord is a single entry in the transaction history.
type TransactionRecord struct {
	Type       string // "deposit" or "withdrawal"
	Amount     int64
	Currency   string
	OccurredAt string // RFC3339
}
