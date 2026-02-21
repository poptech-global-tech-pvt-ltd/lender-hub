package kafka

import (
	"context"

	"github.com/google/uuid"

	"lending-hub-service/config"
	"lending-hub-service/internal/domain/refund/entity"
	refundPort "lending-hub-service/internal/domain/refund/port"
	"lending-hub-service/internal/infrastructure/kafka/events"
	baseLogger "lending-hub-service/pkg/logger"
)

// RefundEventPublisher publishes refund events to lsp.refund.* topics
type RefundEventPublisher struct {
	createdWriter       *TopicWriter
	statusUpdatedWriter *TopicWriter
	logger              *baseLogger.Logger
}

// NewRefundEventPublisher creates a new refund event publisher
func NewRefundEventPublisher(cfg *config.Config, logger *baseLogger.Logger) *RefundEventPublisher {
	brokers := cfg.Kafka.Brokers
	topics := cfg.Kafka.Topics
	prodCfg := cfg.Kafka.Producer

	return &RefundEventPublisher{
		createdWriter:       NewTopicWriter(brokers, topics.RefundCreated, prodCfg, logger),
		statusUpdatedWriter: NewTopicWriter(brokers, topics.RefundStatusUpdated, prodCfg, logger),
		logger:              logger,
	}
}

var _ refundPort.RefundEventPublisher = (*RefundEventPublisher)(nil)

func getStrRefund(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func (p *RefundEventPublisher) PublishRefundCreated(ctx context.Context, refund *entity.Refund) error {
	providerRefundTxnID := ""
	if refund.ProviderRefundTxnID != nil {
		providerRefundTxnID = *refund.ProviderRefundTxnID
	}
	reason := ""
	if refund.Reason != nil {
		reason = string(*refund.Reason)
	}
	evt := events.RefundCreatedEvent{
		BaseEvent:         events.NewBaseEvent("refund.created", uuid.New().String()),
		RefundID:          refund.RefundID,
		PaymentRefundID:   refund.PaymentRefundID,
		ProviderRefundTxnID: providerRefundTxnID,
		PaymentID:         refund.PaymentID,
		LoanID:            refund.LoanID,
		UserID:            refund.UserID,
		Lender:            refund.Lender,
		Amount:            refund.Amount,
		Currency:          refund.Currency,
		Status:            string(refund.Status),
		Reason:            reason,
		CreatedAt:         refund.CreatedAt,
	}
	return p.createdWriter.Publish(ctx, refund.RefundID, evt)
}

func (p *RefundEventPublisher) PublishRefundStatusUpdated(ctx context.Context, refund *entity.Refund, oldStatus entity.RefundStatus, trigger string) error {
	evt := events.RefundStatusUpdatedEvent{
		BaseEvent:       events.NewBaseEvent("refund.status_updated", uuid.New().String()),
		RefundID:        refund.RefundID,
		PaymentRefundID: refund.PaymentRefundID,
		PaymentID:       refund.PaymentID,
		LoanID:          refund.LoanID,
		UserID:          refund.UserID,
		Lender:          refund.Lender,
		OldStatus:       string(oldStatus),
		NewStatus:       string(refund.Status),
		Trigger:         trigger,
		UpdatedAt:       refund.UpdatedAt,
	}
	return p.statusUpdatedWriter.Publish(ctx, refund.RefundID, evt)
}

// Close closes all writers
func (p *RefundEventPublisher) Close() error {
	_ = p.createdWriter.Close()
	_ = p.statusUpdatedWriter.Close()
	return nil
}
