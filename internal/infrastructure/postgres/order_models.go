package postgres

import "time"

// LenderPaymentState maps to lender_payment_state
type LenderPaymentState struct {
	ID                   int64      `gorm:"column:id;primaryKey;autoIncrement"`
	PaymentID            string     `gorm:"column:payment_id;type:text;not null;unique"`
	UserID               string     `gorm:"column:user_id;type:text;not null"`
	MerchantID           string     `gorm:"column:merchant_id;type:text;not null"`
	Lender               string     `gorm:"column:lender;type:text;not null"`
	Amount               float64    `gorm:"column:amount;type:numeric(14,2);not null"`
	Currency             string     `gorm:"column:currency;type:text;not null"`
	Status               string     `gorm:"column:status;type:lender_payment_status;not null"`
	Source               *string    `gorm:"column:source;type:request_source"`
	ReturnURL            *string    `gorm:"column:return_url;type:text"`
	EMIPlan              []byte     `gorm:"column:emi_plan;type:jsonb"`
	LenderOrderID        *string    `gorm:"column:lender_order_id;type:text"`
	LenderMerchantTxnID  *string    `gorm:"column:lender_merchant_txn_id;type:text"`
	LenderLastStatus     *string    `gorm:"column:lender_last_status;type:text"`
	LenderLastTxnID      *string    `gorm:"column:lender_last_txn_id;type:text"`
	LenderLastTxnStatus  *string    `gorm:"column:lender_last_txn_status;type:text"`
	LenderLastTxnMessage *string    `gorm:"column:lender_last_txn_message;type:text"`
	LenderLastTxnTime    *time.Time `gorm:"column:lender_last_txn_time"`
	LastErrorCode        *string    `gorm:"column:last_error_code;type:text"`
	LastErrorMessage     *string    `gorm:"column:last_error_message;type:text"`
	CreatedAt            time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt            time.Time  `gorm:"column:updated_at;not null"`
}

func (LenderPaymentState) TableName() string {
	return "lender_payment_state"
}

// LenderPaymentMapping maps to lender_payment_mapping
type LenderPaymentMapping struct {
	ID                   int64     `gorm:"column:id;primaryKey;autoIncrement"`
	PaymentID            string    `gorm:"column:payment_id;type:text;not null"`
	UserID               string    `gorm:"column:user_id;type:text;not null"`
	Lender               string    `gorm:"column:lender;type:text;not null"`
	LenderMerchantTxnID  string    `gorm:"column:lender_merchant_txn_id;type:text;not null"`
	LenderOrderID        *string   `gorm:"column:lender_order_id;type:text"`
	EligibilityResponseID *string  `gorm:"column:eligibility_response_id;type:text"`
	CreatedAt            time.Time `gorm:"column:created_at;not null"`
	UpdatedAt            time.Time `gorm:"column:updated_at;not null"`
}

func (LenderPaymentMapping) TableName() string {
	return "lender_payment_mapping"
}

// LenderIdempotencyKey maps to lender_idempotency_keys
type LenderIdempotencyKey struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement"`
	IdempotencyKey  string    `gorm:"column:idempotency_key;type:text;not null;unique"`
	RequestHash     string    `gorm:"column:request_hash;type:text;not null"`
	Status          string    `gorm:"column:status;type:idempotency_status;not null"`
	ResponsePayload []byte    `gorm:"column:response_payload;type:jsonb"`
	LenderOrderID   *string   `gorm:"column:lender_order_id;type:text"`
	CreatedAt       time.Time `gorm:"column:created_at;not null"`
	ExpiresAt       time.Time `gorm:"column:expires_at;not null"`
}

func (LenderIdempotencyKey) TableName() string {
	return "lender_idempotency_keys"
}
