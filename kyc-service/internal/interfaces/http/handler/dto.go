package handler

// Request DTOs

type SubmitKYCRequest struct {
	CustomerID string `json:"customer_id"`
}

type RejectKYCRequest struct {
	Reason string `json:"reason"`
}

// Response DTOs

type KYCStatusResponse struct {
	VerificationID string `json:"verification_id"`
	CustomerID     string `json:"customer_id"`
	Status         string `json:"status"`
	Reason         string `json:"reason,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}
