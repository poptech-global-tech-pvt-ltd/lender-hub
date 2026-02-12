package entity

import "time"

// Refund represents a refund record
type Refund struct {
	ID            int64
	RefundID      string
	PaymentID     string
	UserID        string
	Lender        string
	Amount        float64
	Currency      string
	Status        RefundStatus
	Reason        *RefundReason
	LenderRefID   *string
	LenderStatus  *string
	LenderMessage *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
