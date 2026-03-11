package port

import "context"

// ProfileEvent types
type ProfileEventType string

const (
	EventProfileActivated ProfileEventType = "ProfileActivated"
	EventProfileBlocked   ProfileEventType = "ProfileBlocked"
	EventProfileUnblocked ProfileEventType = "ProfileUnblocked"
	EventLimitUpdated     ProfileEventType = "LimitUpdated"
	EventStatusChanged    ProfileEventType = "StatusChanged"
)

type ProfileEvent struct {
	Type           ProfileEventType
	UserID         string
	Lender         string
	PreviousStatus string
	NewStatus      string
	CreditLimit    float64
	AvailableLimit float64
	IsBlocked      bool
	BlockReason    string
	BlockSource    string
}

// ProfileEventPublisher publishes profile change events (Kafka in prod, noop stub now)
type ProfileEventPublisher interface {
	Publish(ctx context.Context, event *ProfileEvent) error
}
