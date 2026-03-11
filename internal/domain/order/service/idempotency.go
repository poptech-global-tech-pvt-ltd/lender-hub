package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	req "lending-hub-service/internal/domain/order/dto/request"
	"lending-hub-service/internal/domain/order/entity"
	"lending-hub-service/internal/domain/order/port"
)

// IdempotencyResult represents the result of an idempotency check
type IdempotencyResult string

const (
	IdempotencyNew       IdempotencyResult = "NEW"
	IdempotencyDuplicate IdempotencyResult = "DUPLICATE"
	IdempotencyConflict  IdempotencyResult = "CONFLICT"
	IdempotencyMismatch  IdempotencyResult = "MISMATCH"
)

// IdempotencyService handles idempotency key management
type IdempotencyService struct {
	repo port.IdempotencyRepository
}

// NewIdempotencyService creates a new IdempotencyService
func NewIdempotencyService(repo port.IdempotencyRepository) *IdempotencyService {
	return &IdempotencyService{repo: repo}
}

// Acquire attempts to acquire an idempotency key
func (s *IdempotencyService) Acquire(ctx context.Context, paymentID, requestHash string) (IdempotencyResult, *entity.IdempotencyKey, error) {
	key, err := s.repo.TryAcquire(ctx, paymentID, requestHash)
	if err != nil {
		return "", nil, err
	}

	// Check if this is a new key (just created)
	if key.Status == entity.IdempotencyProcessing {
		// Check if hash matches (should match since we just created it)
		if key.RequestHash == requestHash {
			return IdempotencyNew, key, nil
		}
		// This shouldn't happen, but handle it
		return IdempotencyMismatch, key, nil
	}

	// Key already exists - check hash
	if key.RequestHash != requestHash {
		return IdempotencyMismatch, key, nil
	}

	// Hash matches - check status
	switch key.Status {
	case entity.IdempotencyCompleted:
		return IdempotencyDuplicate, key, nil
	case entity.IdempotencyProcessing:
		return IdempotencyConflict, key, nil
	case entity.IdempotencyFailed:
		// Failed keys can be retried
		return IdempotencyNew, key, nil
	default:
		return IdempotencyConflict, key, nil
	}
}

// Complete marks an idempotency key as completed
func (s *IdempotencyService) Complete(ctx context.Context, paymentID string, response []byte, lenderOrderID *string) error {
	return s.repo.MarkCompleted(ctx, paymentID, response, lenderOrderID)
}

// Fail marks an idempotency key as failed
func (s *IdempotencyService) Fail(ctx context.Context, paymentID string) error {
	return s.repo.MarkFailed(ctx, paymentID)
}

// ComputeHash computes SHA256 hash of the request for idempotency conflict detection
func (s *IdempotencyService) ComputeHash(req req.CreateOrderRequest) string {
	canonical := struct {
		PaymentID string  `json:"paymentId"`
		UserID    string  `json:"userId"`
		Amount    float64 `json:"amount"`
		Currency  string  `json:"currency"`
		Source    string  `json:"source"`
		ReturnURL string  `json:"returnUrl"`
	}{
		PaymentID: req.PaymentID,
		UserID:    req.UserID,
		Amount:    req.Amount,
		Currency:  req.Currency,
		Source:    req.Source,
		ReturnURL: req.ReturnURL,
	}
	jsonBytes, err := json.Marshal(canonical)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(jsonBytes)
	return hex.EncodeToString(hash[:])
}
