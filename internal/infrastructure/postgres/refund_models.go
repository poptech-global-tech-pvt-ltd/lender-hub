package postgres

import "time"

// LenderRefund maps to lender_refunds
type LenderRefund struct {
	ID            int64     `gorm:"column:id;primaryKey;autoIncrement"`
	RefundID      string    `gorm:"column:refund_id;type:text;not null;unique"`
	PaymentID     string    `gorm:"column:payment_id;type:text;not null"`
	UserID        string    `gorm:"column:user_id;type:text;not null"`
	Lender        string    `gorm:"column:lender;type:text;not null"`
	Amount        float64   `gorm:"column:amount;type:numeric(14,2);not null"`
	Currency      string    `gorm:"column:currency;type:text;not null"`
	Status        string    `gorm:"column:status;type:lender_refund_status;not null"`
	Reason        *string   `gorm:"column:reason;type:refund_reason"`
	LenderRefID   *string   `gorm:"column:lender_ref_id;type:text"`
	LenderStatus  *string   `gorm:"column:lender_status;type:text"`
	LenderMessage *string   `gorm:"column:lender_message;type:text"`
	CreatedAt     time.Time `gorm:"column:created_at;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null"`
}

func (LenderRefund) TableName() string {
	return "lender_refunds"
}
