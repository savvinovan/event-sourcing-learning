package kyc

import "github.com/savvinovan/kyc-service/internal/application/command"

const (
	CmdSubmitKYC  command.CommandType = "SubmitKYC"
	CmdApproveKYC command.CommandType = "ApproveKYC"
	CmdRejectKYC  command.CommandType = "RejectKYC"
)

type SubmitKYCCommand struct {
	VerificationID string
	CustomerID     string
}

func (c SubmitKYCCommand) CommandType() command.CommandType { return CmdSubmitKYC }

type ApproveKYCCommand struct {
	VerificationID string
}

func (c ApproveKYCCommand) CommandType() command.CommandType { return CmdApproveKYC }

type RejectKYCCommand struct {
	VerificationID string
	Reason         string
}

func (c RejectKYCCommand) CommandType() command.CommandType { return CmdRejectKYC }
