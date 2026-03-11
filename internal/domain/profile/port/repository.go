package port

import (
	"context"

	"lending-hub-service/internal/domain/profile/entity"
)

// ProfileRepository abstracts persistence of user profile data
type ProfileRepository interface {
	GetByUserIDAndLender(ctx context.Context, userID, lender string) (*entity.UserProfile, error)
	GetForUpdate(ctx context.Context, userID, lender string) (*entity.UserProfile, error)
	Upsert(ctx context.Context, profile *entity.UserProfile) error

	UpsertFromEligibility(ctx context.Context, userID, lender string, result *EligibilityResult) error
	UpsertFromCustomerStatus(ctx context.Context, userID, lender string, result *CustomerStatusResult) error
	UpsertFromCombined(ctx context.Context, userID, lender string, cs *CustomerStatusResult, el *EligibilityResult) error
}
