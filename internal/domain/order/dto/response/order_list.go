package response

import "time"

// OrderSummary is a single item in the list response
type OrderSummary struct {
	LoanID        string    `json:"loanId,omitempty"`       // our ID (lps_xxx)
	PaymentID     string    `json:"paymentId"`              // POP's ID
	Status        string    `json:"status"`                 // NEVER empty
	LenderOrderID string    `json:"lenderOrderId,omitempty"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// OrderListResponse wraps paginated order list
type OrderListResponse struct {
	Orders  []OrderSummary `json:"orders"`
	Total   int            `json:"total"`
	Page    int            `json:"page"`
	PerPage int            `json:"perPage"`
}
