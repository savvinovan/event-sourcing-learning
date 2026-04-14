package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"

	appkyc "github.com/savvinovan/kyc-service/internal/application/kyc"
	"github.com/savvinovan/kyc-service/internal/application/command"
	"github.com/savvinovan/kyc-service/internal/application/query"
	domain "github.com/savvinovan/kyc-service/internal/domain/kyc"
	"github.com/savvinovan/kyc-service/internal/interfaces/http/gen"
)

// compile-time check
var _ gen.StrictServerInterface = (*KYCHandler)(nil)

type KYCHandler struct {
	commands command.Bus
	queries  query.Bus
	log      *slog.Logger
}

func NewKYCHandler(commands command.Bus, queries query.Bus, log *slog.Logger) *KYCHandler {
	return &KYCHandler{commands: commands, queries: queries, log: log}
}

func (h *KYCHandler) SubmitKYC(ctx context.Context, req gen.SubmitKYCRequestObject) (gen.SubmitKYCResponseObject, error) {
	verificationID := domain.NewVerificationID()
	cmd := appkyc.SubmitKYCCommand{
		VerificationID: verificationID,
		CustomerID:     domain.CustomerID(req.Body.CustomerId.String()),
	}
	if err := h.commands.Dispatch(ctx, cmd); err != nil {
		return h.submitErr(err), nil
	}

	id, err := uuid.Parse(string(verificationID))
	if err != nil {
		h.log.Error("SubmitKYC: failed to parse generated verification ID", "error", err)
		return gen.SubmitKYC500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}, nil
	}
	return gen.SubmitKYC201JSONResponse{VerificationId: id}, nil
}

func (h *KYCHandler) submitErr(err error) gen.SubmitKYCResponseObject {
	switch {
	case errors.Is(err, domain.ErrVerificationAlreadyExists):
		return gen.SubmitKYC422JSONResponse{UnprocessableEntityJSONResponse: gen.UnprocessableEntityJSONResponse{Message: err.Error()}}
	default:
		h.log.Error("SubmitKYC unhandled error", "error", err)
		return gen.SubmitKYC500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}
	}
}

func (h *KYCHandler) ApproveKYC(ctx context.Context, req gen.ApproveKYCRequestObject) (gen.ApproveKYCResponseObject, error) {
	cmd := appkyc.ApproveKYCCommand{
		VerificationID: domain.VerificationID(req.Id.String()),
	}
	if err := h.commands.Dispatch(ctx, cmd); err != nil {
		return h.approveErr(err), nil
	}
	return gen.ApproveKYC204Response{}, nil
}

func (h *KYCHandler) approveErr(err error) gen.ApproveKYCResponseObject {
	switch {
	case errors.Is(err, domain.ErrVerificationNotFound):
		return gen.ApproveKYC404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: err.Error()}}
	case errors.Is(err, domain.ErrAlreadyVerified),
		errors.Is(err, domain.ErrAlreadyRejected),
		errors.Is(err, domain.ErrNotSubmitted):
		return gen.ApproveKYC422JSONResponse{UnprocessableEntityJSONResponse: gen.UnprocessableEntityJSONResponse{Message: err.Error()}}
	default:
		h.log.Error("ApproveKYC unhandled error", "error", err)
		return gen.ApproveKYC500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}
	}
}

func (h *KYCHandler) RejectKYC(ctx context.Context, req gen.RejectKYCRequestObject) (gen.RejectKYCResponseObject, error) {
	cmd := appkyc.RejectKYCCommand{
		VerificationID: domain.VerificationID(req.Id.String()),
		Reason:         req.Body.Reason,
	}
	if err := h.commands.Dispatch(ctx, cmd); err != nil {
		return h.rejectErr(err), nil
	}
	return gen.RejectKYC204Response{}, nil
}

func (h *KYCHandler) rejectErr(err error) gen.RejectKYCResponseObject {
	switch {
	case errors.Is(err, domain.ErrVerificationNotFound):
		return gen.RejectKYC404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: err.Error()}}
	case errors.Is(err, domain.ErrAlreadyVerified),
		errors.Is(err, domain.ErrAlreadyRejected),
		errors.Is(err, domain.ErrNotSubmitted):
		return gen.RejectKYC422JSONResponse{UnprocessableEntityJSONResponse: gen.UnprocessableEntityJSONResponse{Message: err.Error()}}
	default:
		h.log.Error("RejectKYC unhandled error", "error", err)
		return gen.RejectKYC500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}
	}
}

func (h *KYCHandler) GetKYCStatus(ctx context.Context, req gen.GetKYCStatusRequestObject) (gen.GetKYCStatusResponseObject, error) {
	result, err := h.queries.Ask(ctx, appkyc.GetKYCStatusQuery{VerificationID: domain.VerificationID(req.Id.String())})
	if err != nil {
		return h.getStatusErr(err), nil
	}

	status, ok := result.(appkyc.KYCStatusResult)
	if !ok {
		h.log.Error("GetKYCStatus: unexpected query result type")
		return gen.GetKYCStatus500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}, nil
	}

	verificationID, err := uuid.Parse(string(status.VerificationID))
	if err != nil {
		h.log.Error("GetKYCStatus: failed to parse verification ID", "error", err)
		return gen.GetKYCStatus500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}, nil
	}
	customerID, err := uuid.Parse(string(status.CustomerID))
	if err != nil {
		h.log.Error("GetKYCStatus: failed to parse customer ID", "error", err)
		return gen.GetKYCStatus500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}, nil
	}

	resp := gen.GetKYCStatus200JSONResponse{
		VerificationId: verificationID,
		CustomerId:     customerID,
		Status:         gen.KYCStatusResponseStatus(status.Status),
	}
	if status.Reason != "" {
		resp.Reason = &status.Reason
	}

	return resp, nil
}

func (h *KYCHandler) getStatusErr(err error) gen.GetKYCStatusResponseObject {
	switch {
	case errors.Is(err, domain.ErrVerificationNotFound):
		return gen.GetKYCStatus404JSONResponse{NotFoundJSONResponse: gen.NotFoundJSONResponse{Message: err.Error()}}
	default:
		h.log.Error("GetKYCStatus unhandled error", "error", err)
		return gen.GetKYCStatus500JSONResponse{InternalErrorJSONResponse: gen.InternalErrorJSONResponse{Message: "internal error"}}
	}
}
