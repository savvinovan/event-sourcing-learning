package account

import (
	"context"
	"fmt"

	domain "github.com/savvinovan/wallet-service/internal/domain/account"
	"github.com/savvinovan/wallet-service/internal/infrastructure/eventstore"
)

// expectedVersion returns the version before the aggregate's uncommitted changes.
func expectedVersion(agg *domain.Account) int {
	return agg.Version() - len(agg.Changes())
}

// --- Command Handlers ---

type OpenAccountHandler struct{ store eventstore.EventStore }
type DepositMoneyHandler struct{ store eventstore.EventStore }
type WithdrawMoneyHandler struct{ store eventstore.EventStore }
type ActivateAccountHandler struct{ store eventstore.EventStore }
type FreezeAccountHandler struct{ store eventstore.EventStore }

func NewOpenAccountHandler(s eventstore.EventStore) *OpenAccountHandler       { return &OpenAccountHandler{s} }
func NewDepositMoneyHandler(s eventstore.EventStore) *DepositMoneyHandler     { return &DepositMoneyHandler{s} }
func NewWithdrawMoneyHandler(s eventstore.EventStore) *WithdrawMoneyHandler   { return &WithdrawMoneyHandler{s} }
func NewActivateAccountHandler(s eventstore.EventStore) *ActivateAccountHandler { return &ActivateAccountHandler{s} }
func NewFreezeAccountHandler(s eventstore.EventStore) *FreezeAccountHandler   { return &FreezeAccountHandler{s} }

func (h *OpenAccountHandler) Handle(ctx context.Context, cmd OpenAccountCommand) error {
	events, err := h.store.Load(ctx, cmd.AccountID)
	if err != nil {
		return fmt.Errorf("open account: load: %w", err)
	}
	agg := &domain.Account{}
	agg.Restore(events)

	if err := agg.Open(cmd.AccountID, cmd.CustomerID, cmd.Currency); err != nil {
		return err
	}
	if err := h.store.Append(ctx, cmd.AccountID, agg.Changes(), expectedVersion(agg)); err != nil {
		return fmt.Errorf("open account: append: %w", err)
	}
	agg.ClearChanges()
	return nil
}

func (h *DepositMoneyHandler) Handle(ctx context.Context, cmd DepositMoneyCommand) error {
	events, err := h.store.Load(ctx, cmd.AccountID)
	if err != nil {
		return fmt.Errorf("deposit: load: %w", err)
	}
	if len(events) == 0 {
		return domain.ErrAccountNotFound
	}
	agg := &domain.Account{}
	agg.Restore(events)

	if err := agg.Deposit(cmd.Amount, cmd.Currency); err != nil {
		return err
	}
	if err := h.store.Append(ctx, cmd.AccountID, agg.Changes(), expectedVersion(agg)); err != nil {
		return fmt.Errorf("deposit: append: %w", err)
	}
	agg.ClearChanges()
	return nil
}

func (h *WithdrawMoneyHandler) Handle(ctx context.Context, cmd WithdrawMoneyCommand) error {
	events, err := h.store.Load(ctx, cmd.AccountID)
	if err != nil {
		return fmt.Errorf("withdraw: load: %w", err)
	}
	if len(events) == 0 {
		return domain.ErrAccountNotFound
	}
	agg := &domain.Account{}
	agg.Restore(events)

	if err := agg.Withdraw(cmd.Amount, cmd.Currency); err != nil {
		return err
	}
	if err := h.store.Append(ctx, cmd.AccountID, agg.Changes(), expectedVersion(agg)); err != nil {
		return fmt.Errorf("withdraw: append: %w", err)
	}
	agg.ClearChanges()
	return nil
}

func (h *ActivateAccountHandler) Handle(ctx context.Context, cmd ActivateAccountCommand) error {
	events, err := h.store.Load(ctx, cmd.AccountID)
	if err != nil {
		return fmt.Errorf("activate: load: %w", err)
	}
	if len(events) == 0 {
		return domain.ErrAccountNotFound
	}
	agg := &domain.Account{}
	agg.Restore(events)

	if err := agg.Activate(); err != nil {
		return err
	}
	if err := h.store.Append(ctx, cmd.AccountID, agg.Changes(), expectedVersion(agg)); err != nil {
		return fmt.Errorf("activate: append: %w", err)
	}
	agg.ClearChanges()
	return nil
}

func (h *FreezeAccountHandler) Handle(ctx context.Context, cmd FreezeAccountCommand) error {
	events, err := h.store.Load(ctx, cmd.AccountID)
	if err != nil {
		return fmt.Errorf("freeze: load: %w", err)
	}
	if len(events) == 0 {
		return domain.ErrAccountNotFound
	}
	agg := &domain.Account{}
	agg.Restore(events)

	if err := agg.Freeze(cmd.Reason); err != nil {
		return err
	}
	if err := h.store.Append(ctx, cmd.AccountID, agg.Changes(), expectedVersion(agg)); err != nil {
		return fmt.Errorf("freeze: append: %w", err)
	}
	agg.ClearChanges()
	return nil
}
