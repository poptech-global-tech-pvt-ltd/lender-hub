package stub

import (
	"context"

	"lending-hub-service/internal/domain/refund/entity"
	"lending-hub-service/internal/domain/refund/port"
)

// StubRefundEventPublisher implements port.RefundEventPublisher as a no-op for tests
type StubRefundEventPublisher struct{}

// NewStubRefundEventPublisher creates a new stub
func NewStubRefundEventPublisher() port.RefundEventPublisher {
	return &StubRefundEventPublisher{}
}

func (p *StubRefundEventPublisher) PublishRefundCreated(ctx context.Context, refund *entity.Refund) error {
	return nil
}

func (p *StubRefundEventPublisher) PublishRefundStatusUpdated(ctx context.Context, refund *entity.Refund, oldStatus entity.RefundStatus, trigger string) error {
	return nil
}
