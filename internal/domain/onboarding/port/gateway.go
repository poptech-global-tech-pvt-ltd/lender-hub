package port

import (
	"context"

	res "lending-hub-service/internal/domain/onboarding/dto/response"
)

// OnboardingGateway abstracts external onboarding provider calls
type OnboardingGateway interface {
	StartOnboarding(ctx context.Context, mobile, email string) (*res.OnboardingResponse, error)
	GetOnboardingStatus(ctx context.Context, onboardingID string) (*res.OnboardingStatusResponse, error)
}
