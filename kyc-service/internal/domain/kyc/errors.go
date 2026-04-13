package kyc

import "errors"

var (
	ErrVerificationAlreadyExists = errors.New("kyc: verification already exists")
	ErrVerificationNotFound      = errors.New("kyc: verification not found")
	ErrNotSubmitted              = errors.New("kyc: must be in Submitted status to approve or reject")
	ErrAlreadyVerified           = errors.New("kyc: verification is already verified")
	ErrAlreadyRejected           = errors.New("kyc: verification is already rejected")
)
