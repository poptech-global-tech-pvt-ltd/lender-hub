package events

import "time"

// BaseEvent is embedded in all Kafka event payloads.
// Version enables forward-compatible schema evolution.
type BaseEvent struct {
	EventID    string    `json:"eventId"`
	EventType  string    `json:"eventType"`
	Version    string    `json:"version"`
	OccurredAt time.Time `json:"occurredAt"`
	Source     string    `json:"source"`
}

// NewBaseEvent creates a base event with standard fields
func NewBaseEvent(eventType string, eventID string) BaseEvent {
	return BaseEvent{
		EventID:    eventID,
		EventType:  eventType,
		Version:    "v1",
		OccurredAt: time.Now().UTC(),
		Source:     "lending-hub-service",
	}
}
