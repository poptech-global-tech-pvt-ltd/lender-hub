package entity

import "time"

// PaymentMapping represents a mapping between payment ID and lender transaction IDs
type PaymentMapping struct {
	ID                    int64
	PaymentID             string
	UserID                string
	Lender                string
	LenderMerchantTxnID   string
	LenderOrderID         *string
	EligibilityResponseID *string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
