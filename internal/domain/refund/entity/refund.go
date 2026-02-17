package entity

import "time"

// Refund represents a refund record
type Refund struct {
	ID                     int64
	RefundID               string
	PaymentID              string
	UserID                 string
	Lender                 string
	Amount                 float64
	Currency               string
	Status                 RefundStatus
	Reason                 *RefundReason
	
	// Provider mapping
	ProviderMerchantTxnID  *string
	ProviderParentTxnID    *string
	ProviderRefundTxnID    *string
	ProviderRefundRefID    string // idempotency key
	
	// Provider status
	LenderRefID            *string
	LenderStatus           *string
	LenderMessage          *string
	LastEnquiredAt         *time.Time
	
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// MarkSuccess updates refund to SUCCESS state with provider transaction IDs
func (r *Refund) MarkSuccess(lpTxnID, parentTxnID, message string) {
	r.Status = RefundStatusSuccess
	if lpTxnID != "" {
		r.ProviderRefundTxnID = &lpTxnID
	}
	if parentTxnID != "" {
		r.ProviderParentTxnID = &parentTxnID
	}
	msg := message
	r.LenderStatus = &msg
	if message != "" {
		r.LenderMessage = &message
	}
	r.UpdatedAt = time.Now().UTC()
}

// MarkFailed updates refund to FAILED state
func (r *Refund) MarkFailed(lenderStatus, message string) {
	r.Status = RefundStatusFailed
	if lenderStatus != "" {
		r.LenderStatus = &lenderStatus
	}
	if message != "" {
		r.LenderMessage = &message
	}
	r.UpdatedAt = time.Now().UTC()
}

// MarkUnknown updates refund to UNKNOWN state (timeout)
func (r *Refund) MarkUnknown(message string) {
	r.Status = RefundStatusUnknown
	status := "TIMEOUT"
	r.LenderStatus = &status
	if message != "" {
		r.LenderMessage = &message
	}
	r.UpdatedAt = time.Now().UTC()
}

// MarkProcessing updates refund to PROCESSING state
func (r *Refund) MarkProcessing(message string) {
	r.Status = RefundStatusProcessing
	if message != "" {
		r.LenderMessage = &message
	}
	r.UpdatedAt = time.Now().UTC()
}

// RecordEnquiry records that an enquiry was performed
func (r *Refund) RecordEnquiry() {
	now := time.Now().UTC()
	r.LastEnquiredAt = &now
	r.UpdatedAt = now
}
