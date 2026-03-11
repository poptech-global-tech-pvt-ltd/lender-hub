package service

import (
	"math/rand"
	"time"

	"lending-hub-service/internal/domain/onboarding/entity"
)

// EventProcessor processes raw callback events and determines status updates
type EventProcessor struct{}

// NewEventProcessor creates a new event processor
func NewEventProcessor() *EventProcessor {
	return &EventProcessor{}
}

// ProcessEvent determines the new onboarding status based on event
func (p *EventProcessor) ProcessEvent(event *entity.OnboardingEvent, currentStatus entity.OnboardingStatus) entity.OnboardingStatus {
	// If already terminal, don't change
	if currentStatus.IsTerminal() {
		return currentStatus
	}

	// Map event status to onboarding status
	switch event.Status {
	case "SUCCESS":
		if event.Step != nil && *event.Step == entity.StepAutopay {
			return entity.OnboardingSuccess
		}
		return entity.OnboardingInProgress
	case "FAILED":
		// Check if error code indicates ineligibility
		if event.ErrorCode != nil {
			classification, found := ClassifyError(*event.ErrorCode)
			if found && classification.Status == entity.OnboardingIneligible {
				return entity.OnboardingIneligible
			}
		}
		return entity.OnboardingFailed
	case "IN_PROGRESS":
		return entity.OnboardingInProgress
	default:
		return currentStatus
	}
}

// CalculateNextRetryTime calculates the next retry time with exponential backoff and jitter
func (p *EventProcessor) CalculateNextRetryTime(errorCode *string, retryCount int) *time.Time {
	if errorCode == nil {
		return nil
	}

	classification, found := ClassifyError(*errorCode)
	if !found || !classification.IsRetryable {
		return nil
	}

	if retryCount >= classification.MaxRetries {
		return nil // Max retries exceeded
	}

	// Exponential backoff: initialDelay * 2^retryCount
	delay := classification.InitialDelay
	for i := 0; i < retryCount; i++ {
		delay *= 2
		if delay > classification.MaxDelay {
			delay = classification.MaxDelay
			break
		}
	}

	// Add jitter: ±20% of delay
	jitter := time.Duration(float64(delay) * 0.2 * (rand.Float64()*2 - 1))
	nextRetry := time.Now().UTC().Add(delay + jitter)

	return &nextRetry
}

// IsRetryable checks if an error is retryable based on error code
func (p *EventProcessor) IsRetryable(errorCode *string) bool {
	if errorCode == nil {
		return false
	}
	classification, found := ClassifyError(*errorCode)
	return found && classification.IsRetryable
}
