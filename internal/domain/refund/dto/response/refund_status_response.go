package response

import "time"

// RefundStatusResponse represents the response for GET refund status endpoint
type RefundStatusResponse struct {
	RefundID       string     `json:"refundId"`
	PaymentID      string     `json:"paymentId"`
	Status         string     `json:"status"`
	Amount         float64    `json:"amount"`
	Currency       string     `json:"currency"`
	Reason         string     `json:"reason,omitempty"`
	LenderStatus   string     `json:"lenderStatus,omitempty"`
	LenderMessage  string     `json:"lenderMessage,omitempty"`
	LastEnquiredAt *time.Time `json:"lastEnquiredAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}
