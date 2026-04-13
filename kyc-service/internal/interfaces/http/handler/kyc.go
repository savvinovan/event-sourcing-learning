package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	appkyc "github.com/savvinovan/kyc-service/internal/application/kyc"
	"github.com/savvinovan/kyc-service/internal/application/command"
	"github.com/savvinovan/kyc-service/internal/application/query"
	domain "github.com/savvinovan/kyc-service/internal/domain/kyc"
)

type KYCHandler struct {
	commands command.Bus
	queries  query.Bus
	log      *slog.Logger
}

func NewKYCHandler(commands command.Bus, queries query.Bus, log *slog.Logger) *KYCHandler {
	return &KYCHandler{commands: commands, queries: queries, log: log}
}

// POST /kyc
func (h *KYCHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var req SubmitKYCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	verificationID := domain.NewVerificationID()
	cmd := appkyc.SubmitKYCCommand{
		VerificationID: verificationID,
		CustomerID:     req.CustomerID,
	}
	if err := h.commands.Dispatch(r.Context(), cmd); err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"verification_id": verificationID})
}

// POST /kyc/{id}/approve
func (h *KYCHandler) Approve(w http.ResponseWriter, r *http.Request) {
	verificationID := chi.URLParam(r, "id")

	cmd := appkyc.ApproveKYCCommand{VerificationID: verificationID}
	if err := h.commands.Dispatch(r.Context(), cmd); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /kyc/{id}/reject
func (h *KYCHandler) Reject(w http.ResponseWriter, r *http.Request) {
	verificationID := chi.URLParam(r, "id")

	var req RejectKYCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cmd := appkyc.RejectKYCCommand{
		VerificationID: verificationID,
		Reason:         req.Reason,
	}
	if err := h.commands.Dispatch(r.Context(), cmd); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /kyc/{id}
func (h *KYCHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	verificationID := chi.URLParam(r, "id")

	result, err := h.queries.Ask(r.Context(), appkyc.GetKYCStatusQuery{VerificationID: verificationID})
	if err != nil {
		h.handleError(w, err)
		return
	}

	status, ok := result.(appkyc.KYCStatusResult)
	if !ok {
		h.log.Error("unexpected query result type")
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := KYCStatusResponse{
		VerificationID: status.VerificationID,
		CustomerID:     status.CustomerID,
		Status:         status.Status,
		Reason:         status.Reason,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *KYCHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrVerificationNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrAlreadyVerified),
		errors.Is(err, domain.ErrAlreadyRejected),
		errors.Is(err, domain.ErrNotSubmitted),
		errors.Is(err, domain.ErrVerificationAlreadyExists):
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
