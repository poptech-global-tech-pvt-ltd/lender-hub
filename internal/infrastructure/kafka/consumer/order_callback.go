package consumer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"

	"lending-hub-service/internal/domain/order/dto/request"
	kafkapkg "lending-hub-service/internal/infrastructure/kafka"
	"lending-hub-service/internal/infrastructure/kafka/events"
	"lending-hub-service/internal/infrastructure/observability/metrics"
	"lending-hub-service/pkg/lender"
	baseLogger "lending-hub-service/pkg/logger"

	"go.uber.org/zap"
)

// OrderCallbackProcessor processes order callbacks from Kafka
type OrderCallbackProcessor interface {
	ProcessCallback(ctx context.Context, req request.OrderCallbackRequest) error
}

// OrderCallbackConsumer consumes from lsp.lazypay.order.callback
type OrderCallbackConsumer struct {
	processor   OrderCallbackProcessor
	dlqWriter   *kafkapkg.TopicWriter
	mc          metrics.MetricsClient
	logger      *baseLogger.Logger
	maxRetries  int
}

// NewOrderCallbackConsumer creates a new order callback consumer
func NewOrderCallbackConsumer(
	processor OrderCallbackProcessor,
	dlqWriter *kafkapkg.TopicWriter,
	mc metrics.MetricsClient,
	logger *baseLogger.Logger,
	maxRetries int,
) *OrderCallbackConsumer {
	if maxRetries < 1 {
		maxRetries = 3
	}
	return &OrderCallbackConsumer{
		processor:  processor,
		dlqWriter:  dlqWriter,
		mc:         mc,
		logger:     logger,
		maxRetries: maxRetries,
	}
}

// Handle processes a single Kafka message
func (c *OrderCallbackConsumer) Handle(ctx context.Context, msg kafka.Message) error {
	var event events.OrderCallbackEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		c.logger.Error("failed to parse OrderCallbackEvent",
			zap.String("raw", string(msg.Value)),
			zap.Error(err))
		c.sendToDLQ(ctx, msg, "PARSE_ERROR", err)
		return nil // commit — skip bad message
	}

	req := c.mapEventToRequest(&event)

	var lastErr error
	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		err := c.processor.ProcessCallback(ctx, req)
		if err == nil {
			c.mc.Count("lsp.kafka.order_callback.processed", 1, []string{"status:success"})
			return nil
		}
		lastErr = err
		c.logger.Warn("ProcessCallback failed, will retry",
			zap.String("loanId", event.LoanID),
			zap.String("paymentId", event.PaymentID),
			zap.Int("attempt", attempt),
			zap.Error(err))
		// Exponential backoff: 100ms, 400ms, 900ms
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Duration(attempt*attempt) * 100 * time.Millisecond):
		}
	}

	c.sendToDLQ(ctx, msg, "MAX_RETRIES_EXCEEDED", lastErr)
	c.mc.Count("lsp.kafka.order_callback.dlq", 1, nil)
	return nil // commit — message in DLQ
}

func (c *OrderCallbackConsumer) mapEventToRequest(e *events.OrderCallbackEvent) request.OrderCallbackRequest {
	status := e.LenderTxnStatus
	if status == "" {
		status = e.LenderStatus
	}
	req := request.OrderCallbackRequest{
		PaymentID:     e.PaymentID,
		Provider:      lender.Lazypay.String(),
		Status:        status,
		EventTime:     e.EventTime.Format(time.RFC3339),
		ErrorMessage:  ptrOrNil(e.LenderTxnMessage),
	}
	if e.LenderOrderID != "" {
		req.LenderOrderID = &e.LenderOrderID
	}
	if e.LenderTxnID != "" {
		req.LenderTxnID = &e.LenderTxnID
	}
	return req
}

func ptrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (c *OrderCallbackConsumer) sendToDLQ(ctx context.Context, original kafka.Message, reason string, err error) {
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
