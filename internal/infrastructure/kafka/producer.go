package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer wraps Kafka message publishing
type Producer struct {
	writer *kafka.Writer
	config ProducerConfig
}

// Verify EventPublisher interface compliance
var _ EventPublisher = (*Producer)(nil)

// NewProducer creates a new Kafka producer
func NewProducer(cfg ProducerConfig) (*Producer, error) {
	// Map compression type - kafka-go uses Codec interface
	// For now, we'll set it to nil (no compression) and add compression later if needed
	// kafka-go handles compression differently - it's set per message or via transport
	var compression kafka.Compression = 0 // 0 = no compression
	switch cfg.CompressionType {
	case "snappy":
		compression = 2 // Snappy
	case "gzip":
		compression = 1 // Gzip
	default:
		compression = 0 // None
	}

	// Map required acks
	var requiredAcks kafka.RequiredAcks
	switch cfg.RequiredAcks {
	case "all":
		requiredAcks = kafka.RequireAll
	case "leader":
		requiredAcks = kafka.RequireOne
	default:
		requiredAcks = kafka.RequireNone
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.LeastBytes{},
		Compression:  compression,
		BatchSize:    cfg.BatchSize,
		BatchTimeout: time.Duration(cfg.LingerMs) * time.Millisecond,
		RequiredAcks: requiredAcks,
		Async:        cfg.Async,
		WriteTimeout: 10 * time.Second,
	}

	return &Producer{
		writer: writer,
		config: cfg,
	}, nil
}

// Publish sends a message to the specified topic
// key = userID or paymentID for partitioning
func (p *Producer) Publish(ctx context.Context, topic, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: data,
	}
	return p.writer.WriteMessages(ctx, msg)
}

// Close gracefully shuts down the producer
func (p *Producer) Close() error {
	return p.writer.Close()
}
