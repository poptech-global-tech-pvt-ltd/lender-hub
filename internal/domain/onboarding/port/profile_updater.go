package port

import "context"

// ProfileUpdater is a narrow interface owned by the onboarding module.
// Only the methods onboarding actually calls are included here.
type ProfileUpdater interface {
	// UpdateOnOnboardingSuccess sets up the credit limit for a user after onboarding
	UpdateOnOnboardingSuccess(ctx context.Context, userID, lender string, amount float64) error
}
