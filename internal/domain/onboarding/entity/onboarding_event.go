package entity

import "time"

// OnboardingEvent represents an event in the onboarding process
type OnboardingEvent struct {
	ID           int64
	Provider     string
	MerchantID   string
	UserID       string
	Mobile       string
	OnboardingID string
	EventType    string
	Status       string
	Step         *OnboardingStep
	ErrorCode    *string
	Message      *string
	EventTime    time.Time
	RawPayload   []byte
	CreatedAt    time.Time
}
