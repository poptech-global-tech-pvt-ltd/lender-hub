package port

import (
	"context"

	"lending-hub-service/internal/domain/onboarding/entity"
)

// OnboardingRepository manages onboarding projection state
type OnboardingRepository interface {
	Create(ctx context.Context, ob *entity.Onboarding) error
	GetByOnboardingID(ctx context.Context, onboardingID string) (*entity.Onboarding, error)
	GetForUpdate(ctx context.Context, onboardingID string) (*entity.Onboarding, error)
	Update(ctx context.Context, ob *entity.Onboarding) error
}

// OnboardingEventStore manages append-only event log
type OnboardingEventStore interface {
	Append(ctx context.Context, event *entity.OnboardingEvent) error
	ListByOnboardingID(ctx context.Context, onboardingID string) ([]*entity.OnboardingEvent, error)
}
