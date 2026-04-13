package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	appaccount "github.com/savvinovan/wallet-service/internal/application/account"
	"github.com/savvinovan/wallet-service/internal/application/command"
	"github.com/savvinovan/wallet-service/internal/application/query"
	domain "github.com/savvinovan/wallet-service/internal/domain/account"
)

type AccountHandler struct {
	commands command.Bus
	queries  query.Bus
	log      *slog.Logger
}

func NewAccountHandler(commands command.Bus, queries query.Bus, log *slog.Logger) *AccountHandler {
	return &AccountHandler{commands: commands, queries: queries, log: log}
}

// POST /accounts
func (h *AccountHandler) OpenAccount(w http.ResponseWriter, r *http.Request) {
	var req OpenAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	accountID := domain.NewAccountID()
	cmd := appaccount.OpenAccountCommand{
		AccountID:  accountID,
		CustomerID: domain.CustomerID(req.CustomerID),
		Currency:   req.Currency,
	}
	if err := h.commands.Dispatch(r.Context(), cmd); err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"account_id": string(accountID)})
}

// POST /accounts/{id}/deposit
func (h *AccountHandler) Deposit(w http.ResponseWriter, r *http.Request) {
	accountID := domain.AccountID(chi.URLParam(r, "id"))

	var req DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := appaccount.DepositMoneyCommand{
		AccountID: accountID,
		Amount:    req.Amount,
		Currency:  req.Currency,
	}
	if err := h.commands.Dispatch(r.Context(), cmd); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /accounts/{id}/withdraw
func (h *AccountHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	accountID := domain.AccountID(chi.URLParam(r, "id"))

	var req WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := appaccount.WithdrawMoneyCommand{
		AccountID: accountID,
		Amount:    req.Amount,
		Currency:  req.Currency,
	}
	if err := h.commands.Dispatch(r.Context(), cmd); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /accounts/{id}/balance
func (h *AccountHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	accountID := domain.AccountID(chi.URLParam(r, "id"))

	result, err := h.queries.Ask(r.Context(), appaccount.GetBalanceQuery{AccountID: accountID})
	if err != nil {
		h.handleError(w, err)
		return
	}

	balance, ok := result.(appaccount.BalanceResult)
	if !ok {
		h.log.Error("unexpected query result type")
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := BalanceResponse{
		AccountID:  string(balance.AccountID),
		CustomerID: string(balance.CustomerID),
		Balance:    balance.Balance,
		Currency:   balance.Currency,
		Status:     balance.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// GET /accounts/{id}/transactions
func (h *AccountHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	accountID := domain.AccountID(chi.URLParam(r, "id"))

	result, err := h.queries.Ask(r.Context(), appaccount.GetTransactionsQuery{AccountID: accountID})
	if err != nil {
		h.handleError(w, err)
		return
	}

	records, ok := result.([]appaccount.TransactionRecord)
	if !ok {
		h.log.Error("unexpected query result type")
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := make([]TransactionResponse, len(records))
	for i, rec := range records {
		resp[i] = TransactionResponse{
			Type:       rec.Type,
			Amount:     rec.Amount,
			Currency:   rec.Currency,
			OccurredAt: rec.OccurredAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *AccountHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrAccountNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrInsufficientFunds),
		errors.Is(err, domain.ErrNotActive),
		errors.Is(err, domain.ErrNotPending),
		errors.Is(err, domain.ErrAccountAlreadyExists),
		errors.Is(err, domain.ErrCurrencyMismatch),
		errors.Is(err, domain.ErrNonPositiveAmount):
		writeError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		h.log.Error("unhandled error", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Message: msg})
}
