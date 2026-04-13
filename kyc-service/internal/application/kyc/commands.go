package kyc

import (
	"github.com/savvinovan/kyc-service/internal/application/command"
	domain "github.com/savvinovan/kyc-service/internal/domain/kyc"
)

const (
	CmdSubmitKYC  command.CommandType = "SubmitKYC"
	CmdApproveKYC command.CommandType = "ApproveKYC"
	CmdRejectKYC  command.CommandType = "RejectKYC"
)

type SubmitKYCCommand struct {
	VerificationID domain.VerificationID
	CustomerID     domain.CustomerID
}

func (c SubmitKYCCommand) CommandType() command.CommandType { return CmdSubmitKYC }

type ApproveKYCCommand struct {
	VerificationID domain.VerificationID
}

func (c ApproveKYCCommand) CommandType() command.CommandType { return CmdApproveKYC }

type RejectKYCCommand struct {
	VerificationID domain.VerificationID
	Reason         string
}

func (c RejectKYCCommand) CommandType() command.CommandType { return CmdRejectKYC }
