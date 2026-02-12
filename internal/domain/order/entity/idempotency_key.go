package entity

import "time"

type IdempotencyStatus string

const (
	IdempotencyProcessing IdempotencyStatus = "PROCESSING"
	IdempotencyCompleted  IdempotencyStatus = "COMPLETED"
	IdempotencyFailed     IdempotencyStatus = "FAILED"
)

// IdempotencyKey represents an idempotency key record
type IdempotencyKey struct {
	ID              int64
	Key             string // = paymentId
	RequestHash     string // SHA256 of request body
	Status          IdempotencyStatus
	ResponsePayload []byte
	LenderOrderID   *string
	CreatedAt       time.Time
	ExpiresAt       time.Time
}
