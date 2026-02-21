package contact

import (
	"context"

	onboardingPort "lending-hub-service/internal/domain/onboarding/port"
	profileService "lending-hub-service/internal/domain/profile/service"
)

// OnboardingContactAdapter adapts profile.UserContactResolver to onboarding.ContactResolver.
// Keeps onboarding module free of profile module dependencies.
type OnboardingContactAdapter struct {
	inner *profileService.UserContactResolver
}

// NewOnboardingContactAdapter creates an adapter.
func NewOnboardingContactAdapter(inner *profileService.UserContactResolver) *OnboardingContactAdapter {
	return &OnboardingContactAdapter{inner: inner}
}

// GetContact implements onboardingPort.ContactResolver.
func (a *OnboardingContactAdapter) GetContact(ctx context.Context, userID string) (*onboardingPort.ContactInfo, error) {
	uc, err := a.inner.Resolve(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &onboardingPort.ContactInfo{Mobile: uc.Mobile, Email: uc.Email}, nil
}
