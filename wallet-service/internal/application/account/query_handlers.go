package account

import (
	"context"
	"fmt"
)

// GetBalanceHandler reads account state from the read model — no event replay.
type GetBalanceHandler struct{ repo AccountReadRepository }

func NewGetBalanceHandler(repo AccountReadRepository) *GetBalanceHandler {
	return &GetBalanceHandler{repo}
}

func (h *GetBalanceHandler) Handle(ctx context.Context, q GetBalanceQuery) (BalanceResult, error) {
	result, err := h.repo.GetBalance(ctx, q.AccountID)
	if err != nil {
		return BalanceResult{}, fmt.Errorf("get balance: %w", err)
	}
	return result, nil
}

// GetTransactionsHandler reads transaction history from the read model — no event replay.
type GetTransactionsHandler struct{ repo AccountReadRepository }

func NewGetTransactionsHandler(repo AccountReadRepository) *GetTransactionsHandler {
	return &GetTransactionsHandler{repo}
}

func (h *GetTransactionsHandler) Handle(ctx context.Context, q GetTransactionsQuery) ([]TransactionRecord, error) {
	records, err := h.repo.GetTransactions(ctx, q.AccountID)
	if err != nil {
		return nil, fmt.Errorf("get transactions: %w", err)
	}
	// Nil slice means account exists but has no transactions — return empty, not nil.
	if records == nil {
		return []TransactionRecord{}, nil
	}
	return records, nil
}
