package kafka

import (
	"context"
	"log"
)

// EventPublisher is the interface both Producer and NoopProducer satisfy
type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, value interface{}) error
	Close() error
}

// NoopProducer implements the same Publish interface but does nothing
// Used when Kafka is not configured (local dev)
type NoopProducer struct{}

// NewNoopProducer creates a new noop producer
func NewNoopProducer() *NoopProducer {
	return &NoopProducer{}
}

// Publish logs at DEBUG level but does nothing
func (p *NoopProducer) Publish(ctx context.Context, topic, key string, value interface{}) error {
	// Log at DEBUG: "Noop publish: topic=%s key=%s"
	log.Printf("[DEBUG] Noop publish: topic=%s key=%s", topic, key)
	return nil
}

// Close does nothing
func (p *NoopProducer) Close() error {
	return nil
}
