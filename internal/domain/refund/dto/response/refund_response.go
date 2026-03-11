package response

import "time"

// RefundResponse used by POST and all GET single-refund endpoints
type RefundResponse struct {
	RefundID            string     `json:"refundId"`
	PaymentRefundID     string     `json:"paymentRefundId"`
	ProviderRefundTxnID string     `json:"providerRefundTxnId,omitempty"`
	PaymentID           string     `json:"paymentId"`
	LoanID              string     `json:"loanId"`
	Status              string     `json:"status"`
	Amount              float64    `json:"amount"`
	Currency            string     `json:"currency"`
	Reason              string     `json:"reason,omitempty"`
	LenderStatus        string     `json:"lenderStatus,omitempty"`
	LenderMessage       string     `json:"lenderMessage,omitempty"`
	LastEnquiredAt      *time.Time `json:"lastEnquiredAt,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}
