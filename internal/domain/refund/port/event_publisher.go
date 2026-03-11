package port

import (
	"context"

	"lending-hub-service/internal/domain/refund/entity"
)

// RefundEventPublisher publishes refund change events to Kafka
type RefundEventPublisher interface {
	PublishRefundCreated(ctx context.Context, refund *entity.Refund) error
	PublishRefundStatusUpdated(ctx context.Context, refund *entity.Refund, oldStatus entity.RefundStatus, trigger string) error
}
