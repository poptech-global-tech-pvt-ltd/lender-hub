package port

import "context"

// OrderEventType represents types of order events
type OrderEventType string

const (
	EventOrderCreated  OrderEventType = "OrderCreated"
	EventOrderCompleted OrderEventType = "OrderCompleted"
	EventOrderFailed    OrderEventType = "OrderFailed"
	EventOrderRefunded  OrderEventType = "OrderRefunded"
)

// OrderEvent represents an order domain event
type OrderEvent struct {
	Type            OrderEventType
	PaymentID       string
	UserID          string
	MerchantID      string
	Lender          string
	Amount          float64
	Currency        string
	Status          string
	LenderOrderID   *string
	LenderTxnID     *string
	ErrorCode       *string
	ErrorMessage    *string
}

// OrderEventPublisher publishes order change events (Kafka in prod, noop stub now)
type OrderEventPublisher interface {
	Publish(ctx context.Context, event *OrderEvent) error
}
