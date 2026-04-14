package handler

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"

	appaccount "github.com/savvinovan/wallet-service/internal/application/account"
	"github.com/savvinovan/wallet-service/internal/application/command"
	"github.com/savvinovan/wallet-service/internal/application/query"
	domain "github.com/savvinovan/wallet-service/internal/domain/account"
	"github.com/savvinovan/wallet-service/internal/interfaces/http/gen"
)

// compile-time check
var _ gen.StrictServerInterface = (*AccountHandler)(nil)

type AccountHandler struct {
	commands command.Bus
	queries  query.Bus
	log      *slog.Logger
}

func NewAccountHandler(commands command.Bus, queries query.Bus, log *slog.Logger) *AccountHandler {
	return &AccountHandler{commands: commands, queries: queries, log: log}
}

func (h *AccountHandler) OpenAccount(ctx context.Context, req gen.OpenAccountRequestObject) (gen.OpenAccountResponseObject, error) {
	accountID := domain.NewAccountID()
	cmd := appaccount.OpenAccountCommand{
		AccountID:  accountID,
		CustomerID: domain.CustomerID(req.Body.CustomerId.String()),
		Currency:   req.Body.Currency,
	}
	if err := h.commands.Dispatch(ctx, cmd); err != nil {
		return h.openAccountErr(err), nil
	}

	id, err := uuid.Parse(string(accountID))
	if err != nil {
		h.log.Error("failed to parse generated account ID", "error", err)
		return gen.OpenAccount500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}, nil
	}
	return gen.OpenAccount201JSONResponse{AccountId: id}, nil
}

func (h *AccountHandler) openAccountErr(err error) gen.OpenAccountResponseObject {
	switch {
	case errors.Is(err, domain.ErrAccountAlreadyExists):
		return gen.OpenAccount422JSONResponse{UnprocessableEntityJSONResponse: gen.UnprocessableEntityJSONResponse{Message: err.Error()}}
	default:
		h.log.Error("OpenAccount unhandled error", "error", err)
		return gen.OpenAccount500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}
	}
}

func (h *AccountHandler) Deposit(ctx context.Context, req gen.DepositRequestObject) (gen.DepositResponseObject, error) {
	money, err := domain.NewMoney(req.Body.Amount, req.Body.Currency)
	if err != nil {
		return gen.Deposit422JSONResponse{UnprocessableEntityJSONResponse: gen.UnprocessableEntityJSONResponse{Message: err.Error()}}, nil
	}
	cmd := appaccount.DepositMoneyCommand{
		AccountID: domain.AccountID(req.Id.String()),
		Amount:    money,
	}
	if err := h.commands.Dispatch(ctx, cmd); err != nil {
		return h.depositErr(err), nil
	}
	return gen.Deposit204Response{}, nil
}

func (h *AccountHandler) depositErr(err error) gen.DepositResponseObject {
	switch {
	case errors.Is(err, domain.ErrAccountNotFound):
		return gen.Deposit404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: err.Error()}}
	case errors.Is(err, domain.ErrInsufficientFunds),
		errors.Is(err, domain.ErrNotActive),
		errors.Is(err, domain.ErrCurrencyMismatch),
		errors.Is(err, domain.ErrNonPositiveAmount):
		return gen.Deposit422JSONResponse{UnprocessableEntityJSONResponse: gen.UnprocessableEntityJSONResponse{Message: err.Error()}}
	default:
		h.log.Error("Deposit unhandled error", "error", err)
		return gen.Deposit500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}
	}
}

func (h *AccountHandler) Withdraw(ctx context.Context, req gen.WithdrawRequestObject) (gen.WithdrawResponseObject, error) {
	money, err := domain.NewMoney(req.Body.Amount, req.Body.Currency)
	if err != nil {
		return gen.Withdraw422JSONResponse{UnprocessableEntityJSONResponse: gen.UnprocessableEntityJSONResponse{Message: err.Error()}}, nil
	}
	cmd := appaccount.WithdrawMoneyCommand{
		AccountID: domain.AccountID(req.Id.String()),
		Amount:    money,
	}
	if err := h.commands.Dispatch(ctx, cmd); err != nil {
		return h.withdrawErr(err), nil
	}
	return gen.Withdraw204Response{}, nil
}

func (h *AccountHandler) withdrawErr(err error) gen.WithdrawResponseObject {
	switch {
	case errors.Is(err, domain.ErrAccountNotFound):
		return gen.Withdraw404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: err.Error()}}
	case errors.Is(err, domain.ErrInsufficientFunds),
		errors.Is(err, domain.ErrNotActive),
		errors.Is(err, domain.ErrCurrencyMismatch),
		errors.Is(err, domain.ErrNonPositiveAmount):
		return gen.Withdraw422JSONResponse{UnprocessableEntityJSONResponse: gen.UnprocessableEntityJSONResponse{Message: err.Error()}}
	default:
		h.log.Error("Withdraw unhandled error", "error", err)
		return gen.Withdraw500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}
	}
}

func (h *AccountHandler) GetBalance(ctx context.Context, req gen.GetBalanceRequestObject) (gen.GetBalanceResponseObject, error) {
	result, err := h.queries.Ask(ctx, appaccount.GetBalanceQuery{AccountID: domain.AccountID(req.Id.String())})
	if err != nil {
		return h.getBalanceErr(err), nil
	}

	balance, ok := result.(appaccount.BalanceResult)
	if !ok {
		h.log.Error("GetBalance: unexpected query result type")
		return gen.GetBalance500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}, nil
	}

	accountID, err := uuid.Parse(string(balance.AccountID))
	if err != nil {
		h.log.Error("GetBalance: failed to parse account ID", "error", err)
		return gen.GetBalance500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}, nil
	}
	customerID, err := uuid.Parse(string(balance.CustomerID))
	if err != nil {
		h.log.Error("GetBalance: failed to parse customer ID", "error", err)
		return gen.GetBalance500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}, nil
	}

	return gen.GetBalance200JSONResponse{
		AccountId:  accountID,
		CustomerId: customerID,
		Balance:    balance.Balance.Amount,
		Currency:   balance.Balance.Currency,
		Status:     gen.BalanceResponseStatus(balance.Status),
	}, nil
}

func (h *AccountHandler) getBalanceErr(err error) gen.GetBalanceResponseObject {
	switch {
	case errors.Is(err, domain.ErrAccountNotFound):
		return gen.GetBalance404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: err.Error()}}
	default:
		h.log.Error("GetBalance unhandled error", "error", err)
		return gen.GetBalance500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}
	}
}

func (h *AccountHandler) GetTransactions(ctx context.Context, req gen.GetTransactionsRequestObject) (gen.GetTransactionsResponseObject, error) {
	result, err := h.queries.Ask(ctx, appaccount.GetTransactionsQuery{AccountID: domain.AccountID(req.Id.String())})
	if err != nil {
		return h.getTransactionsErr(err), nil
	}

	records, ok := result.([]appaccount.TransactionRecord)
	if !ok {
		h.log.Error("GetTransactions: unexpected query result type")
		return gen.GetTransactions500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}, nil
	}

	resp := make(gen.GetTransactions200JSONResponse, len(records))
	for i, rec := range records {
		t, err := time.Parse(time.RFC3339, rec.OccurredAt)
		if err != nil {
			h.log.Error("GetTransactions: failed to parse occurred_at", "error", err, "value", rec.OccurredAt)
			return gen.GetTransactions500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}, nil
		}
		resp[i] = gen.TransactionRecord{
			Type:       gen.TransactionRecordType(rec.Type),
			Amount:     rec.Amount.Amount,
			Currency:   rec.Amount.Currency,
			OccurredAt: t,
		}
	}

	return resp, nil
}

func (h *AccountHandler) getTransactionsErr(err error) gen.GetTransactionsResponseObject {
	switch {
	case errors.Is(err, domain.ErrAccountNotFound):
		return gen.GetTransactions404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: err.Error()}}
	default:
		h.log.Error("GetTransactions unhandled error", "error", err)
		return gen.GetTransactions500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}
	}
}

func (h *AccountHandler) ActivateAccount(ctx context.Context, req gen.ActivateAccountRequestObject) (gen.ActivateAccountResponseObject, error) {
	cmd := appaccount.ActivateAccountCommand{
		AccountID: domain.AccountID(req.Id.String()),
	}
	if err := h.commands.Dispatch(ctx, cmd); err != nil {
		return h.activateErr(err), nil
	}
	return gen.ActivateAccount204Response{}, nil
}

func (h *AccountHandler) activateErr(err error) gen.ActivateAccountResponseObject {
	switch {
	case errors.Is(err, domain.ErrAccountNotFound):
		return gen.ActivateAccount404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: err.Error()}}
	case errors.Is(err, domain.ErrNotPending):
		return gen.ActivateAccount422JSONResponse{UnprocessableEntityJSONResponse: gen.UnprocessableEntityJSONResponse{Message: err.Error()}}
	default:
		h.log.Error("ActivateAccount unhandled error", "error", err)
		return gen.ActivateAccount500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}
	}
}

func (h *AccountHandler) FreezeAccount(ctx context.Context, req gen.FreezeAccountRequestObject) (gen.FreezeAccountResponseObject, error) {
	cmd := appaccount.FreezeAccountCommand{
		AccountID: domain.AccountID(req.Id.String()),
		Reason:    req.Body.Reason,
	}
	if err := h.commands.Dispatch(ctx, cmd); err != nil {
		return h.freezeErr(err), nil
	}
	return gen.FreezeAccount204Response{}, nil
}

func (h *AccountHandler) freezeErr(err error) gen.FreezeAccountResponseObject {
	switch {
	case errors.Is(err, domain.ErrAccountNotFound):
		return gen.FreezeAccount404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: err.Error()}}
	case errors.Is(err, domain.ErrNotPending):
		return gen.FreezeAccount422JSONResponse{UnprocessableEntityJSONResponse: gen.UnprocessableEntityJSONResponse{Message: err.Error()}}
	default:
		h.log.Error("FreezeAccount unhandled error", "error", err)
		return gen.FreezeAccount500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}
	}
}
