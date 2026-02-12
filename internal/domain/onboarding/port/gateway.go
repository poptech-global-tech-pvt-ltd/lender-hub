package port

import (
	"context"

	req "lending-hub-service/internal/domain/onboarding/dto/request"
	res "lending-hub-service/internal/domain/onboarding/dto/response"
)

// OnboardingGateway abstracts external onboarding provider calls
type OnboardingGateway interface {
	StartOnboarding(ctx context.Context, req req.StartOnboardingRequest) (*res.OnboardingResponse, error)
	GetOnboardingStatus(ctx context.Context, mobile string) (*res.OnboardingStatusResponse, error)
}
