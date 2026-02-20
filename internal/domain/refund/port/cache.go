package port

import (
	"context"
	"time"

	"lending-hub-service/internal/domain/refund/entity"
)

// RefundCache abstracts short-TTL caching for refund status
// Cache key: "refund:{refundId}"
// TTL: 30s for non-terminal (PENDING/UNKNOWN/PROCESSING), 5min for terminal (SUCCESS/FAILED)
type RefundCache interface {
	Get(ctx context.Context, refundID string) (entity.RefundStatus, bool, error)
	Set(ctx context.Context, refundID string, status entity.RefundStatus, ttl time.Duration) error
	Invalidate(ctx context.Context, refundID string) error
}
