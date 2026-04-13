package handler

// Request DTOs

type OpenAccountRequest struct {
	CustomerID string `json:"customer_id"`
	Currency   string `json:"currency"`
}

type DepositRequest struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type WithdrawRequest struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

// Response DTOs

type BalanceResponse struct {
	AccountID  string `json:"account_id"`
	CustomerID string `json:"customer_id"`
	Balance    int64  `json:"balance"`
	Currency   string `json:"currency"`
	Status     string `json:"status"`
}

type TransactionResponse struct {
	Type       string `json:"type"`
	Amount     int64  `json:"amount"`
	Currency   string `json:"currency"`
	OccurredAt string `json:"occurred_at"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}
