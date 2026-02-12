package entity

import "time"

// Order represents a payment/order state
type Order struct {
	ID                   int64
	PaymentID            string
	UserID               string
	MerchantID           string
	Lender               string
	Amount               float64
	Currency             string
	Status               OrderStatus
	Source               *string
	ReturnURL            *string
	EMIPlan              []byte // JSONB
	LenderOrderID        *string
	LenderMerchantTxnID  *string
	LenderLastStatus     *string
	LenderLastTxnID      *string
	LenderLastTxnStatus  *string
	LenderLastTxnMessage *string
	LenderLastTxnTime    *time.Time
	LastErrorCode        *string
	LastErrorMessage     *string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
