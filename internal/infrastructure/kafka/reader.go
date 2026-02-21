package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"

	"lending-hub-service/config"
	baseLogger "lending-hub-service/pkg/logger"

	"go.uber.org/zap"
)

// TopicMessageHandler processes a single Kafka message; return error to avoid commit (retry)
type TopicMessageHandler func(ctx context.Context, msg kafka.Message) error

// TopicReader wraps kafka-go reader for consuming from a topic
type TopicReader struct {
	r      *kafka.Reader
	logger *baseLogger.Logger
}

// NewTopicReader creates a new topic-specific Kafka reader
func NewTopicReader(brokers []string, topic, groupID string, cfg config.KafkaConsumerConfig, logger *baseLogger.Logger) *TopicReader {
	minBytes := cfg.MinBytes
	if minBytes < 1 {
		minBytes = 1
	}
	maxBytes := cfg.MaxBytes
	if maxBytes < 1 {
		maxBytes = 10 * 1024 * 1024 // 10MB
	}
	maxWait := time.Duration(cfg.MaxWaitSeconds) * time.Second
	if maxWait == 0 {
		maxWait = 3 * time.Second
	}
	commitInterval := time.Duration(cfg.CommitIntervalMs) * time.Millisecond
	if commitInterval == 0 {
		commitInterval = time.Second
	}

	return &TopicReader{
		r: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			Topic:          topic,
			GroupID:        groupID,
			MinBytes:       minBytes,
			MaxBytes:       maxBytes,
			MaxWait:        maxWait,
			CommitInterval: commitInterval,
		}),
		logger: logger,
	}
}

// Consume runs the consume loop; returns on context cancel (graceful shutdown)
func (r *TopicReader) Consume(ctx context.Context, handler TopicMessageHandler) error {
	for {
		msg, err := r.r.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil // graceful shutdown
			}
			return fmt.Errorf("kafka.TopicReader.Consume fetch: %w", err)
		}
		if err := handler(ctx, msg); err != nil {
			r.logger.Warn("message handler failed, not committing",
				zap.String("topic", msg.Topic),
				zap.Int64("offset", msg.Offset),
				zap.Error(err))
			continue
		}
		if err := r.r.CommitMessages(ctx, msg); err != nil {
			r.logger.Warn("commit failed",
				zap.String("topic", msg.Topic),
				zap.Error(err))
		}
	}
}

// Close closes the reader
func (r *TopicReader) Close() error {
	return r.r.Close()
}
