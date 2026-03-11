package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

// MessageHandler processes a single Kafka message
type MessageHandler func(ctx context.Context, topic string, key, value []byte) error

// Consumer wraps Kafka message consumption
type Consumer struct {
	reader  *kafka.Reader
	config  ConsumerConfig
	handler MessageHandler
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg ConsumerConfig, topics []string, handler MessageHandler) (*Consumer, error) {
	// Map offset reset
	var startOffset int64
	switch cfg.OffsetReset {
	case "earliest":
		startOffset = kafka.FirstOffset
	case "latest":
		startOffset = kafka.LastOffset
	default:
		startOffset = kafka.FirstOffset
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		GroupID:     cfg.GroupID,
		GroupTopics: topics,
		StartOffset: startOffset,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
	})

	return &Consumer{
		reader:  reader,
		config:  cfg,
		handler: handler,
	}, nil
}

// Start begins consuming messages in a blocking loop
// Call in a goroutine: go consumer.Start(ctx)
func (c *Consumer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				continue
			}
			if err := c.handler(ctx, msg.Topic, msg.Key, msg.Value); err != nil {
				// Log error, continue (at-least-once)
				log.Printf("Error handling message: topic=%s key=%s error=%v", msg.Topic, string(msg.Key), err)
			}
		}
	}
}

// Close gracefully shuts down the consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}
