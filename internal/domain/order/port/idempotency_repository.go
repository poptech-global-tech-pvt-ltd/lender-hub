package port

import (
	"context"

	"lending-hub-service/internal/domain/order/entity"
)

// IdempotencyRepository manages idempotency keys for order creation
type IdempotencyRepository interface {
	// TryAcquire attempts to create a PROCESSING record; returns existing record if conflict
	TryAcquire(ctx context.Context, key, requestHash string) (*entity.IdempotencyKey, error)

	// MarkCompleted updates status to COMPLETED and stores response payload
	MarkCompleted(ctx context.Context, key string, responsePayload []byte, lenderOrderID *string) error

	// MarkFailed updates status to FAILED
	MarkFailed(ctx context.Context, key string) error

	// Get returns current record for given key
	Get(ctx context.Context, key string) (*entity.IdempotencyKey, error)
}
