package port

import (
	"context"

	"lending-hub-service/internal/domain/profile/entity"
)

// ProfileRepository abstracts persistence of user profile data
type ProfileRepository interface {
	// Get returns a user profile by userId+lender, or nil if not found
	Get(ctx context.Context, userID, lender string) (*entity.UserProfile, error)

	// GetForUpdate returns a user profile row locked for update (FOR UPDATE)
	GetForUpdate(ctx context.Context, userID, lender string) (*entity.UserProfile, error)

	// Upsert creates or updates a profile row
	Upsert(ctx context.Context, profile *entity.UserProfile) error
}
