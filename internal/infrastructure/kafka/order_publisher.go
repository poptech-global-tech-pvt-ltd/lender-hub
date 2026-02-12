package kafka

import (
	"context"
	"time"

	"github.com/google/uuid"

	orderPort "lending-hub-service/internal/domain/order/port"
)

// getStringValue safely extracts string from pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// OrderEventPublisher implements orderPort.OrderEventPublisher via Kafka
type OrderEventPublisher struct {
	publisher EventPublisher
}

// NewOrderEventPublisher creates a new order event publisher
func NewOrderEventPublisher(pub EventPublisher) *OrderEventPublisher {
	return &OrderEventPublisher{publisher: pub}
}

// Verify interface compliance
var _ orderPort.OrderEventPublisher = (*OrderEventPublisher)(nil)

// Publish publishes an order event to Kafka
func (p *OrderEventPublisher) Publish(ctx context.Context, event *orderPort.OrderEvent) error {
	domainEvent := DomainEvent{
		EventID:    uuid.New().String(),
		EventType:  string(event.Type),
		Source:     "payin3-service",
		OccurredAt: time.Now().UTC(),
		UserID:     event.UserID,
		Lender:     event.Lender,
		Data: OrderEventData{
			PaymentID:     event.PaymentID,
			Amount:        event.Amount,
			Currency:      event.Currency,
			Status:        event.Status,
			LenderOrderID: getStringValue(event.LenderOrderID),
			ErrorCode:     getStringValue(event.ErrorCode),
		},
	}
	// Use paymentID as key for partitioning
	return p.publisher.Publish(ctx, TopicOrderEvents, event.PaymentID, domainEvent)
}
