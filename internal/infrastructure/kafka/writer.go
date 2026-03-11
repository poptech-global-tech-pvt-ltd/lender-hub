package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"

	"lending-hub-service/config"
	baseLogger "lending-hub-service/pkg/logger"
)

// TopicWriter wraps kafka-go writer for producing messages to a specific topic
type TopicWriter struct {
	w      *kafka.Writer
	logger *baseLogger.Logger
}

// NewTopicWriter creates a new topic-specific Kafka writer
func NewTopicWriter(brokers []string, topic string, cfg config.KafkaProducerConfig, logger *baseLogger.Logger) *TopicWriter {
	writeTimeout := time.Duration(cfg.WriteTimeoutSeconds) * time.Second
	if writeTimeout == 0 {
		writeTimeout = 10 * time.Second
	}
	batchSize := cfg.BatchSize
	if batchSize < 1 {
		batchSize = 1
	}
	requiredAcks := kafka.RequiredAcks(cfg.RequiredAcks)
	if cfg.RequiredAcks == 0 {
		requiredAcks = kafka.RequireAll
	}

	return &TopicWriter{
		w: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			WriteTimeout: writeTimeout,
			BatchSize:    batchSize,
			RequiredAcks: requiredAcks,
		},
		logger: logger,
	}
}

// Publish serializes value to JSON and writes to Kafka
func (w *TopicWriter) Publish(ctx context.Context, key string, value interface{}) error {
	b, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("kafka.TopicWriter.Publish marshal: %w", err)
	}
	return w.w.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: b,
	})
}

// Close closes the writer
func (w *TopicWriter) Close() error {
	return w.w.Close()
}
