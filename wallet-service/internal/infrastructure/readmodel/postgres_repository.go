package readmodel

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	appaccount "github.com/savvinovan/wallet-service/internal/application/account"
	domain "github.com/savvinovan/wallet-service/internal/domain/account"
)

// PostgresReadModelRepository queries account_read_models and transaction_history.
type PostgresReadModelRepository struct {
	db *pgxpool.Pool
}

func NewPostgresReadModelRepository(db *pgxpool.Pool) *PostgresReadModelRepository {
	return &PostgresReadModelRepository{db: db}
}

func (r *PostgresReadModelRepository) GetBalance(ctx context.Context, accountID domain.AccountID) (appaccount.BalanceResult, error) {
	var (
		accID      string
		customerID string
		status     string
		balanceStr string
		currency   string
		version    int
	)
	err := r.db.QueryRow(ctx, `
		SELECT account_id::text, customer_id::text, status, balance::text, currency, version
		FROM account_read_models
		WHERE account_id = $1::uuid
	`, string(accountID)).Scan(&accID, &customerID, &status, &balanceStr, &currency, &version)
	if err != nil {
		if err == pgx.ErrNoRows {
			return appaccount.BalanceResult{}, domain.ErrAccountNotFound
		}
		return appaccount.BalanceResult{}, fmt.Errorf("get balance: %w", err)
	}

	amount, err := decimal.NewFromString(balanceStr)
	if err != nil {
		return appaccount.BalanceResult{}, fmt.Errorf("get balance: parse amount %q: %w", balanceStr, err)
	}

	return appaccount.BalanceResult{
		AccountID:  domain.AccountID(accID),
		CustomerID: domain.CustomerID(customerID),
		Balance:    domain.Money{Amount: amount, Currency: currency},
		Status:     status,
	}, nil
}

func (r *PostgresReadModelRepository) GetTransactions(ctx context.Context, accountID domain.AccountID) ([]appaccount.TransactionRecord, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM account_read_models WHERE account_id = $1::uuid)`,
		string(accountID),
	).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("get transactions: check account: %w", err)
	}
	if !exists {
		return nil, domain.ErrAccountNotFound
	}

	rows, err := r.db.Query(ctx, `
		SELECT tx_type, amount::text, currency, occurred_at
		FROM transaction_history
		WHERE account_id = $1::uuid
		ORDER BY occurred_at ASC
	`, string(accountID))
	if err != nil {
		return nil, fmt.Errorf("get transactions: query: %w", err)
	}
	defer rows.Close()

	var records []appaccount.TransactionRecord
	for rows.Next() {
		var (
			txType     string
			amountStr  string
			currency   string
			occurredAt time.Time
		)
		if err := rows.Scan(&txType, &amountStr, &currency, &occurredAt); err != nil {
			return nil, fmt.Errorf("get transactions: scan: %w", err)
		}
		amount, err := decimal.NewFromString(amountStr)
		if err != nil {
			return nil, fmt.Errorf("get transactions: parse amount %q: %w", amountStr, err)
		}
		records = append(records, appaccount.TransactionRecord{
			Type:       txType,
			Amount:     domain.Money{Amount: amount, Currency: currency},
			OccurredAt: occurredAt.UTC().Format(time.RFC3339),
		})
	}
	return records, rows.Err()
}

var _ appaccount.AccountReadRepository = (*PostgresReadModelRepository)(nil)
