package events

import "time"

// OrderCreatedEvent — published to lsp.order.created
type OrderCreatedEvent struct {
	BaseEvent
	LoanID        string    `json:"loanId"`
	PaymentID     string    `json:"paymentId"`
	LenderOrderID string    `json:"lenderOrderId"`
	UserID        string    `json:"userId"`
	MerchantID    string    `json:"merchantId"`
	Lender        string    `json:"lender"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
}

// OrderStatusUpdatedEvent — published to lsp.order.status_updated
type OrderStatusUpdatedEvent struct {
	BaseEvent
	LoanID        string    `json:"loanId"`
	PaymentID     string    `json:"paymentId"`
	LenderOrderID string    `json:"lenderOrderId"`
	UserID        string    `json:"userId"`
	Lender        string    `json:"lender"`
	OldStatus     string    `json:"oldStatus"`
	NewStatus     string    `json:"newStatus"`
	Trigger       string    `json:"trigger"` // "callback" | "enquiry" | "support"
	UpdatedAt     time.Time `json:"updatedAt"`
}

// OrderSupportUpdatedEvent — published to lsp.order.support_updated
type OrderSupportUpdatedEvent struct {
	BaseEvent
	LoanID    string    `json:"loanId"`
	PaymentID string    `json:"paymentId"`
	UserID    string    `json:"userId"`
	OldStatus string    `json:"oldStatus"`
	NewStatus string    `json:"newStatus"`
	Reason    string    `json:"reason"`
	Actor     string    `json:"actor"`
	UpdatedAt time.Time `json:"updatedAt"`
}
