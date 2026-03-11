package stub

import (
	"context"

	"lending-hub-service/internal/domain/order/entity"
	"lending-hub-service/internal/domain/order/port"
)

// StubOrderEventPublisher implements port.OrderEventPublisher as a no-op for tests
type StubOrderEventPublisher struct{}

// NewStubOrderEventPublisher creates a new stub event publisher
func NewStubOrderEventPublisher() port.OrderEventPublisher {
	return &StubOrderEventPublisher{}
}

func (p *StubOrderEventPublisher) PublishOrderCreated(ctx context.Context, order *entity.Order) error {
	return nil
}

func (p *StubOrderEventPublisher) PublishOrderStatusUpdated(ctx context.Context, order *entity.Order, oldStatus entity.OrderStatus, trigger string) error {
	return nil
}

func (p *StubOrderEventPublisher) PublishOrderSupportUpdated(ctx context.Context, order *entity.Order, oldStatus entity.OrderStatus, reason, actor string) error {
	return nil
}
