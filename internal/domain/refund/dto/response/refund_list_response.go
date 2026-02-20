package response

import "time"

// RefundSummary for list endpoints
type RefundSummary struct {
	RefundID            string    `json:"refundId"`
	PaymentRefundID     string    `json:"paymentRefundId"`
	ProviderRefundTxnID string    `json:"providerRefundTxnId,omitempty"`
	PaymentID           string    `json:"paymentId"`
	LoanID              string    `json:"loanId"`
	Status              string    `json:"status"`
	Amount              float64   `json:"amount"`
	Currency            string    `json:"currency"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
}

// RefundListResponse for list by order or by user
type RefundListResponse struct {
	Refunds []RefundSummary `json:"refunds"`
	Total   int             `json:"total"`
	Page    int             `json:"page"`
	PerPage int             `json:"perPage"`
}
