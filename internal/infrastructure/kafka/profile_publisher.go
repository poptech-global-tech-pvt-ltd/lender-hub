package kafka

import (
	"context"
	"time"

	"github.com/google/uuid"

	profilePort "lending-hub-service/internal/domain/profile/port"
)

// ProfileEventPublisher implements profilePort.ProfileEventPublisher via Kafka
type ProfileEventPublisher struct {
	publisher EventPublisher
}

// NewProfileEventPublisher creates a new profile event publisher
func NewProfileEventPublisher(pub EventPublisher) *ProfileEventPublisher {
	return &ProfileEventPublisher{publisher: pub}
}

// Verify interface compliance
var _ profilePort.ProfileEventPublisher = (*ProfileEventPublisher)(nil)

// Publish publishes a profile event to Kafka
func (p *ProfileEventPublisher) Publish(ctx context.Context, event *profilePort.ProfileEvent) error {
	domainEvent := DomainEvent{
		EventID:    uuid.New().String(),
		EventType:  string(event.Type),
		Source:     "payin3-service",
		OccurredAt: time.Now().UTC(),
		UserID:     event.UserID,
		Lender:     event.Lender,
		Data: ProfileEventData{
			PreviousStatus: event.PreviousStatus,
			NewStatus:      event.NewStatus,
			CreditLimit:    event.CreditLimit,
			AvailableLimit: event.AvailableLimit,
			IsBlocked:      event.IsBlocked,
			BlockReason:    event.BlockReason,
			BlockSource:    event.BlockSource,
		},
	}
	// Use userID as key for partitioning
	key := event.UserID + ":" + event.Lender
	return p.publisher.Publish(ctx, TopicProfileEvents, key, domainEvent)
}
