package events

import "time"

// RefundCreatedEvent — published to lsp.refund.created
type RefundCreatedEvent struct {
	BaseEvent
	RefundID            string    `json:"refundId"`
	PaymentRefundID     string    `json:"paymentRefundId"`
	ProviderRefundTxnID string    `json:"providerRefundTxnId,omitempty"`
	PaymentID           string    `json:"paymentId"`
	LoanID              string    `json:"loanId"`
	UserID              string    `json:"userId"`
	Lender              string    `json:"lender"`
	Amount              float64   `json:"amount"`
	Currency            string    `json:"currency"`
	Status              string    `json:"status"`
	Reason              string    `json:"reason"`
	CreatedAt           time.Time `json:"createdAt"`
}

// RefundStatusUpdatedEvent — published to lsp.refund.status_updated
type RefundStatusUpdatedEvent struct {
	BaseEvent
	RefundID        string    `json:"refundId"`
	PaymentRefundID string    `json:"paymentRefundId"`
	PaymentID       string    `json:"paymentId"`
	LoanID          string    `json:"loanId"`
	UserID          string    `json:"userId"`
	Lender          string    `json:"lender"`
	OldStatus       string    `json:"oldStatus"`
	NewStatus       string    `json:"newStatus"`
	Trigger         string    `json:"trigger"` // "callback" | "enquiry"
	UpdatedAt       time.Time `json:"updatedAt"`
}
