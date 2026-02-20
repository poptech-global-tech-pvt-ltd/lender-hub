package response

import "time"

// EmiDetail holds EMI plan details for completed orders
type EmiDetail struct {
	Tenure             int     `json:"tenure"`
	EMI                float64 `json:"emi"`
	InterestRate       float64 `json:"interestRate"`
	Principal          float64 `json:"principal"`
	TotalPayableAmount float64 `json:"totalPayableAmount"`
	FirstEmiDueDate    string  `json:"firstEmiDueDate"`
}

// OrderStatusResponse is returned by GET order (by paymentId, loanId, or lenderOrderId)
type OrderStatusResponse struct {
	LoanID           string     `json:"loanId,omitempty"`       // our ID (lps_xxx)
	PaymentID        string     `json:"paymentId"`              // POP's ID
	Status           string     `json:"status"`                 // NEVER empty: PENDING|SUCCESS|FAILED|REFUNDED|EXPIRED|CANCELLED
	LenderOrderID    string     `json:"lenderOrderId,omitempty"`
	Amount           float64    `json:"amount"`
	Currency         string     `json:"currency"`
	EmiPlan          *EmiDetail `json:"emiPlan,omitempty"`
	LenderLastStatus string     `json:"lenderLastStatus,omitempty"`
	LenderLastMessage string    `json:"lenderLastMessage,omitempty"`
	LastErrorCode    string     `json:"lastErrorCode,omitempty"`
	LastErrorMessage string     `json:"lastErrorMessage,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}
