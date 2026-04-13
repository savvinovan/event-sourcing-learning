package account

import (
	"context"
	"fmt"
	"time"

	domain "github.com/savvinovan/wallet-service/internal/domain/account"
	"github.com/savvinovan/wallet-service/internal/domain/event"
	"github.com/savvinovan/wallet-service/internal/infrastructure/eventstore"
)

// GetBalanceHandler rebuilds account state from the event store and returns the balance.
type GetBalanceHandler struct{ store eventstore.EventStore }

func NewGetBalanceHandler(s eventstore.EventStore) *GetBalanceHandler {
	return &GetBalanceHandler{s}
}

func (h *GetBalanceHandler) Handle(ctx context.Context, q GetBalanceQuery) (BalanceResult, error) {
	events, err := h.store.Load(ctx, q.AccountID)
	if err != nil {
		return BalanceResult{}, fmt.Errorf("get balance: load: %w", err)
	}
	if len(events) == 0 {
		return BalanceResult{}, domain.ErrAccountNotFound
	}
	agg := &domain.Account{}
	agg.Restore(events)

	return BalanceResult{
		AccountID:  agg.ID(),
		CustomerID: agg.CustomerID(),
		Balance:    agg.Balance(),
		Currency:   agg.Currency(),
		Status:     agg.Status().String(),
	}, nil
}

// GetTransactionsHandler replays events and returns deposits and withdrawals.
type GetTransactionsHandler struct{ store eventstore.EventStore }

func NewGetTransactionsHandler(s eventstore.EventStore) *GetTransactionsHandler {
	return &GetTransactionsHandler{s}
}

func (h *GetTransactionsHandler) Handle(ctx context.Context, q GetTransactionsQuery) ([]TransactionRecord, error) {
	events, err := h.store.Load(ctx, q.AccountID)
	if err != nil {
		return nil, fmt.Errorf("get transactions: load: %w", err)
	}
	if len(events) == 0 {
		return nil, domain.ErrAccountNotFound
	}

	var records []TransactionRecord
	for _, e := range events {
		rec, ok := toTransactionRecord(e)
		if ok {
			records = append(records, rec)
		}
	}
	return records, nil
}

func toTransactionRecord(e event.DomainEvent) (TransactionRecord, bool) {
	switch v := e.(type) {
	case domain.MoneyDeposited:
		return TransactionRecord{
			Type:       "deposit",
			Amount:     v.Amount,
			Currency:   v.Currency,
			OccurredAt: v.OccurredAt().Format(time.RFC3339),
		}, true
	case domain.MoneyWithdrawn:
		return TransactionRecord{
			Type:       "withdrawal",
			Amount:     v.Amount,
			Currency:   v.Currency,
			OccurredAt: v.OccurredAt().Format(time.RFC3339),
		}, true
	default:
		return TransactionRecord{}, false
	}
}
