package projector

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	domain "github.com/savvinovan/wallet-service/internal/domain/account"
	"github.com/savvinovan/wallet-service/internal/domain/event"
)

// AccountProjectorName is the checkpoint key used in projector_checkpoints.
const AccountProjectorName = "account_projector"

// eventHandler applies a single domain event to the read model within a transaction.
type eventHandler func(ctx context.Context, tx pgx.Tx, e event.DomainEvent) error

// AccountProjector updates account_read_models and transaction_history.
// Each event type is handled by a dedicated function registered in the handlers map —
// no monolithic switch statement.
type AccountProjector struct {
	handlers map[string]eventHandler
}

func NewAccountProjector() *AccountProjector {
	p := &AccountProjector{handlers: make(map[string]eventHandler)}
	p.handlers[domain.EventTypeAccountOpened] = p.onAccountOpened
	p.handlers[domain.EventTypeMoneyDeposited] = p.onMoneyDeposited
	p.handlers[domain.EventTypeMoneyWithdrawn] = p.onMoneyWithdrawn
	p.handlers[domain.EventTypeAccountActivated] = p.onAccountActivated
	p.handlers[domain.EventTypeAccountFrozen] = p.onAccountFrozen
	return p
}

func (p *AccountProjector) Apply(ctx context.Context, tx pgx.Tx, events []event.DomainEvent) error {
	for _, e := range events {
		h, ok := p.handlers[e.EventType()]
		if !ok {
			continue // unknown event types are silently skipped (forward compatibility)
		}
		if err := h(ctx, tx, e); err != nil {
			return fmt.Errorf("apply %s (seq=%d): %w", e.EventType(), e.Version(), err)
		}
	}
	return nil
}

func (p *AccountProjector) onAccountOpened(ctx context.Context, tx pgx.Tx, e event.DomainEvent) error {
	v := e.(domain.AccountOpened)
	_, err := tx.Exec(ctx, `
		INSERT INTO account_read_models
			(account_id, customer_id, status, balance, currency, version, updated_at)
		VALUES ($1::uuid, $2::uuid, 'Pending', 0, $3, $4, $5)
	`, v.AggregateID(), string(v.CustomerID), v.Currency, v.Version(), v.OccurredAt())
	return err
}

func (p *AccountProjector) onMoneyDeposited(ctx context.Context, tx pgx.Tx, e event.DomainEvent) error {
	v := e.(domain.MoneyDeposited)
	if _, err := tx.Exec(ctx, `
		UPDATE account_read_models
		SET balance = balance + $1::numeric, version = $2, updated_at = $3
		WHERE account_id = $4::uuid
	`, v.Amount.Amount.String(), v.Version(), v.OccurredAt(), v.AggregateID()); err != nil {
		return err
	}
	return p.insertTransaction(ctx, tx, v.AggregateID(), "deposit", v.Amount.Amount.String(), v.Amount.Currency, v.OccurredAt())
}

func (p *AccountProjector) onMoneyWithdrawn(ctx context.Context, tx pgx.Tx, e event.DomainEvent) error {
	v := e.(domain.MoneyWithdrawn)
	if _, err := tx.Exec(ctx, `
		UPDATE account_read_models
		SET balance = balance - $1::numeric, version = $2, updated_at = $3
		WHERE account_id = $4::uuid
	`, v.Amount.Amount.String(), v.Version(), v.OccurredAt(), v.AggregateID()); err != nil {
		return err
	}
	return p.insertTransaction(ctx, tx, v.AggregateID(), "withdrawal", v.Amount.Amount.String(), v.Amount.Currency, v.OccurredAt())
}

func (p *AccountProjector) onAccountActivated(ctx context.Context, tx pgx.Tx, e event.DomainEvent) error {
	_, err := tx.Exec(ctx, `
		UPDATE account_read_models
		SET status = 'Active', version = $1, updated_at = $2
		WHERE account_id = $3::uuid
	`, e.Version(), e.OccurredAt(), e.AggregateID())
	return err
}

func (p *AccountProjector) onAccountFrozen(ctx context.Context, tx pgx.Tx, e event.DomainEvent) error {
	_, err := tx.Exec(ctx, `
		UPDATE account_read_models
		SET status = 'Frozen', version = $1, updated_at = $2
		WHERE account_id = $3::uuid
	`, e.Version(), e.OccurredAt(), e.AggregateID())
	return err
}

func (p *AccountProjector) insertTransaction(ctx context.Context, tx pgx.Tx, accountID, txType, amount, currency string, occurredAt any) error {
	txID := uuid.Must(uuid.NewV7()).String()
	_, err := tx.Exec(ctx, `
		INSERT INTO transaction_history (id, account_id, tx_type, amount, currency, occurred_at)
		VALUES ($1::uuid, $2::uuid, $3, $4::numeric, $5, $6)
	`, txID, accountID, txType, amount, currency, occurredAt)
	return err
}

var _ EventApplier = (*AccountProjector)(nil)
