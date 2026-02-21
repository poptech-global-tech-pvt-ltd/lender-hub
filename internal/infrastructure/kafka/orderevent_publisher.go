package kafka

import (
	"context"

	"github.com/google/uuid"

	"lending-hub-service/config"
	"lending-hub-service/internal/domain/order/entity"
	orderPort "lending-hub-service/internal/domain/order/port"
	"lending-hub-service/internal/infrastructure/kafka/events"
	baseLogger "lending-hub-service/pkg/logger"
)

// OrderEventPublisher publishes order events to lsp.order.* topics
type OrderEventPublisher struct {
	createdWriter       *TopicWriter
	statusUpdatedWriter *TopicWriter
	supportUpdatedWriter *TopicWriter
	logger              *baseLogger.Logger
}

// NewOrderEventPublisher creates a new order event publisher
func NewOrderEventPublisher(cfg *config.Config, logger *baseLogger.Logger) *OrderEventPublisher {
	brokers := cfg.Kafka.Brokers
	topics := cfg.Kafka.Topics
	prodCfg := cfg.Kafka.Producer

	return &OrderEventPublisher{
		createdWriter:        NewTopicWriter(brokers, topics.OrderCreated, prodCfg, logger),
		statusUpdatedWriter:  NewTopicWriter(brokers, topics.OrderStatusUpdated, prodCfg, logger),
		supportUpdatedWriter: NewTopicWriter(brokers, topics.OrderSupportUpdated, prodCfg, logger),
		logger:               logger,
	}
}

// NewOrderEventPublisherWithTopics creates publisher with explicit topic config (for tests/fallback)
func NewOrderEventPublisherWithTopics(brokers []string, topics config.KafkaTopics, prodCfg config.KafkaProducerConfig, logger *baseLogger.Logger) *OrderEventPublisher {
	return &OrderEventPublisher{
		createdWriter:        NewTopicWriter(brokers, topics.OrderCreated, prodCfg, logger),
		statusUpdatedWriter:  NewTopicWriter(brokers, topics.OrderStatusUpdated, prodCfg, logger),
		supportUpdatedWriter: NewTopicWriter(brokers, topics.OrderSupportUpdated, prodCfg, logger),
		logger:               logger,
	}
}

var _ orderPort.OrderEventPublisher = (*OrderEventPublisher)(nil)

func getStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func (p *OrderEventPublisher) PublishOrderCreated(ctx context.Context, order *entity.Order) error {
	loanID := getStr(order.LenderMerchantTxnID)
	lenderOrderID := getStr(order.LenderOrderID)
	evt := events.OrderCreatedEvent{
		BaseEvent:     events.NewBaseEvent("order.created", uuid.New().String()),
		LoanID:        loanID,
		PaymentID:     order.PaymentID,
		LenderOrderID: lenderOrderID,
		UserID:        order.UserID,
		MerchantID:    order.MerchantID,
		Lender:        order.Lender,
		Amount:        order.Amount,
		Currency:      order.Currency,
		Status:        string(order.Status),
		CreatedAt:     order.CreatedAt,
	}
	key := loanID
	if key == "" {
		key = order.PaymentID
	}
	return p.createdWriter.Publish(ctx, key, evt)
}

func (p *OrderEventPublisher) PublishOrderStatusUpdated(ctx context.Context, order *entity.Order, oldStatus entity.OrderStatus, trigger string) error {
	loanID := getStr(order.LenderMerchantTxnID)
	lenderOrderID := getStr(order.LenderOrderID)
	evt := events.OrderStatusUpdatedEvent{
		BaseEvent:     events.NewBaseEvent("order.status_updated", uuid.New().String()),
		LoanID:        loanID,
		PaymentID:     order.PaymentID,
		LenderOrderID: lenderOrderID,
		UserID:        order.UserID,
		Lender:        order.Lender,
		OldStatus:     string(oldStatus),
		NewStatus:     string(order.Status),
		Trigger:       trigger,
		UpdatedAt:     order.UpdatedAt,
	}
	key := loanID
	if key == "" {
		key = order.PaymentID
	}
	return p.statusUpdatedWriter.Publish(ctx, key, evt)
}

func (p *OrderEventPublisher) PublishOrderSupportUpdated(ctx context.Context, order *entity.Order, oldStatus entity.OrderStatus, reason, actor string) error {
	loanID := getStr(order.LenderMerchantTxnID)
	evt := events.OrderSupportUpdatedEvent{
		BaseEvent: events.NewBaseEvent("order.support_updated", uuid.New().String()),
		LoanID:    loanID,
		PaymentID: order.PaymentID,
		UserID:    order.UserID,
		OldStatus: string(oldStatus),
		NewStatus: string(order.Status),
		Reason:    reason,
		Actor:     actor,
		UpdatedAt: order.UpdatedAt,
	}
	key := loanID
	if key == "" {
		key = order.PaymentID
	}
	return p.supportUpdatedWriter.Publish(ctx, key, evt)
}

// Close closes all writers
func (p *OrderEventPublisher) Close() error {
	_ = p.createdWriter.Close()
	_ = p.statusUpdatedWriter.Close()
	_ = p.supportUpdatedWriter.Close()
	return nil
}
