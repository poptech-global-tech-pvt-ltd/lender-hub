package stub

import (
	"context"

	"lending-hub-service/internal/domain/order/port"
)

// StubOrderEventPublisher implements port.OrderEventPublisher as a no-op for local development
type StubOrderEventPublisher struct{}

// NewStubOrderEventPublisher creates a new stub event publisher
func NewStubOrderEventPublisher() port.OrderEventPublisher {
	return &StubOrderEventPublisher{}
}

// Publish is a no-op in stub implementation
func (p *StubOrderEventPublisher) Publish(ctx context.Context, event *port.OrderEvent) error {
	// No-op: in production this would publish to Kafka
	return nil
}
