package stub

import (
	"context"

	"lending-hub-service/internal/domain/profile/port"
)

// StubProfileEventPublisher implements port.ProfileEventPublisher as a no-op for local development
type StubProfileEventPublisher struct{}

// NewStubProfileEventPublisher creates a new stub event publisher
func NewStubProfileEventPublisher() port.ProfileEventPublisher {
	return &StubProfileEventPublisher{}
}

// Publish is a no-op in stub implementation
func (p *StubProfileEventPublisher) Publish(ctx context.Context, event *port.ProfileEvent) error {
	// No-op: in production this would publish to Kafka
	return nil
}
