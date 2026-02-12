package entity

import "time"

// IdempotencyKey represents an idempotency key record
type IdempotencyKey struct {
	ID              int64
	IdempotencyKey  string
	RequestHash     string
	Status          string
	ResponsePayload []byte
	LenderOrderID   *string
	CreatedAt       time.Time
	ExpiresAt       time.Time
}
