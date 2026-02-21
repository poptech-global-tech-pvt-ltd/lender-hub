package consumer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"

	"lending-hub-service/internal/domain/refund/dto/request"
	kafkapkg "lending-hub-service/internal/infrastructure/kafka"
	"lending-hub-service/internal/infrastructure/kafka/events"
	"lending-hub-service/internal/infrastructure/observability/metrics"
	baseLogger "lending-hub-service/pkg/logger"

	"go.uber.org/zap"
)

// RefundCallbackProcessor processes refund callbacks from Kafka
type RefundCallbackProcessor interface {
	ProcessRefundCallback(ctx context.Context, req request.RefundCallbackRequest) error
}

// RefundCallbackConsumer consumes from lsp.lazypay.refund.callback
type RefundCallbackConsumer struct {
	processor  RefundCallbackProcessor
	dlqWriter  *kafkapkg.TopicWriter
	mc         metrics.MetricsClient
	logger     *baseLogger.Logger
	maxRetries int
}

// NewRefundCallbackConsumer creates a new refund callback consumer
func NewRefundCallbackConsumer(
	processor RefundCallbackProcessor,
	dlqWriter *kafkapkg.TopicWriter,
	mc metrics.MetricsClient,
	logger *baseLogger.Logger,
	maxRetries int,
) *RefundCallbackConsumer {
	if maxRetries < 1 {
		maxRetries = 3
	}
	return &RefundCallbackConsumer{
		processor:  processor,
		dlqWriter:  dlqWriter,
		mc:         mc,
		logger:     logger,
		maxRetries: maxRetries,
	}
}

// Handle processes a single Kafka message
func (c *RefundCallbackConsumer) Handle(ctx context.Context, msg kafka.Message) error {
	var event events.RefundCallbackEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		c.logger.Error("failed to parse RefundCallbackEvent",
			zap.String("raw", string(msg.Value)),
			zap.Error(err))
		c.sendToDLQ(ctx, msg, "PARSE_ERROR", err)
		return nil
	}

	req := request.RefundCallbackRequest{
		RefundID:         event.RefundID,
		PaymentRefundID:  event.PaymentRefundID,
		LoanID:           event.LoanID,
		PaymentID:        event.PaymentID,
		LenderStatus:     event.LenderStatus,
		LenderTxnID:      event.LenderTxnID,
		LenderTxnStatus:  event.LenderTxnStatus,
		LenderTxnMessage: event.LenderTxnMessage,
		EventTime:        event.EventTime,
	}

	var lastErr error
	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		err := c.processor.ProcessRefundCallback(ctx, req)
		if err == nil {
			c.mc.Count("lsp.kafka.refund_callback.processed", 1, []string{"status:success"})
			return nil
		}
		lastErr = err
		c.logger.Warn("ProcessRefundCallback failed, will retry",
			zap.String("refundId", event.RefundID),
			zap.String("paymentRefundId", event.PaymentRefundID),
			zap.Int("attempt", attempt),
			zap.Error(err))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt*attempt) * 100 * time.Millisecond):
		}
	}

	c.sendToDLQ(ctx, msg, "MAX_RETRIES_EXCEEDED", lastErr)
	c.mc.Count("lsp.kafka.refund_callback.dlq", 1, nil)
	return nil
}

func (c *RefundCallbackConsumer) sendToDLQ(ctx context.Context, original kafka.Message, reason string, err error) {
	payload := map[string]interface{}{
		"originalTopic":  original.Topic,
		"originalOffset": original.Offset,
		"originalKey":    string(original.Key),
		"originalValue":  string(original.Value),
		"failureReason":  reason,
		"errorMessage":   "",
		"failedAt":       time.Now().UTC(),
	}
	if err != nil {
		payload["errorMessage"] = err.Error()
	}
	b, _ := json.Marshal(payload)
	if publishErr := c.dlqWriter.Publish(ctx, string(original.Key), b); publishErr != nil {
		c.logger.Error("failed to send message to DLQ",
			zap.String("reason", reason),
			zap.Error(publishErr))
	}
}
