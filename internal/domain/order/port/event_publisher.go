package port

import (
	"context"

	"lending-hub-service/internal/domain/order/entity"
)

// OrderEventPublisher publishes order change events to Kafka
type OrderEventPublisher interface {
	PublishOrderCreated(ctx context.Context, order *entity.Order) error
	PublishOrderStatusUpdated(ctx context.Context, order *entity.Order, oldStatus entity.OrderStatus, trigger string) error
	PublishOrderSupportUpdated(ctx context.Context, order *entity.Order, oldStatus entity.OrderStatus, reason, actor string) error
}
