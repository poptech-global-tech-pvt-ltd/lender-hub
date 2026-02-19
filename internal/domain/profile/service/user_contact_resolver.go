package service

import (
	"context"
	"fmt"

	"lending-hub-service/internal/domain/profile/entity"
	"lending-hub-service/internal/domain/profile/repository"
	"lending-hub-service/internal/infrastructure/userprofile"
	baseLogger "lending-hub-service/pkg/logger"
)

// UserContactResolver orchestrates resolution of userId → mobile + email
type UserContactResolver struct {
	repo          *repository.UserContactRepository
	profileClient *userprofile.Client
	logger        *baseLogger.Logger
}

// NewUserContactResolver creates a new UserContactResolver
func NewUserContactResolver(
	repo *repository.UserContactRepository,
	profileClient *userprofile.Client,
	logger *baseLogger.Logger,
) *UserContactResolver {
	return &UserContactResolver{
		repo:          repo,
		profileClient: profileClient,
		logger:        logger,
	}
}

// Resolve returns mobile + email for a given userId.
// Strategy: local DB first → external profile service → upsert → return
// Returns error if mobile cannot be resolved.
func (r *UserContactResolver) Resolve(ctx context.Context, userID string) (*entity.UserContact, error) {
	// 1. Check local cache (lender_user_profile)
	cached, err := r.repo.GetByUserID(ctx, userID)
	if err != nil {
		r.logger.Error("failed to check local user contact",
			baseLogger.UserID(userID),
			baseLogger.ErrorCode(err.Error()),
		)
		// Don't fail — try external
	}

	if cached != nil {
		r.logger.Debug("user contact found in local DB",
			baseLogger.UserID(userID),
		)
		return cached, nil
	}

	// 2. Not in local DB → call external profile service
	r.logger.Info("user contact not cached, calling profile service",
		baseLogger.UserID(userID),
	)

	contactInfo, err := r.profileClient.GetUserContact(ctx, userID)
	if err != nil {
		r.logger.Error("profile service call failed",
			baseLogger.UserID(userID),
			baseLogger.ErrorCode(err.Error()),
		)
		return nil, fmt.Errorf("cannot resolve user contact: %w", err)
	}

	if contactInfo.Mobile == "" {
		return nil, fmt.Errorf("mobile not found for user %s", userID)
	}

	// 3. Upsert into local DB
	uc := &entity.UserContact{
		UserID:   userID,
		Mobile:   contactInfo.Mobile,
		Email:    contactInfo.Email,
		RawPhone: contactInfo.RawPhone,
		Source:   "PROFILE_SERVICE",
	}

	if err := r.repo.Upsert(ctx, uc); err != nil {
		r.logger.Warn("failed to cache user contact locally",
			baseLogger.UserID(userID),
			baseLogger.ErrorCode(err.Error()),
		)
		// Don't fail the request — we have the data, just couldn't cache
	}

	return uc, nil
}

// RefreshFromSource forces a fresh fetch from external profile service
// and updates local DB. Use on onboarding success / order creation.
func (r *UserContactResolver) RefreshFromSource(ctx context.Context, userID, source string) (*entity.UserContact, error) {
	contactInfo, err := r.profileClient.GetUserContact(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("refresh user contact failed: %w", err)
	}

	uc := &entity.UserContact{
		UserID:   userID,
		Mobile:   contactInfo.Mobile,
		Email:    contactInfo.Email,
		RawPhone: contactInfo.RawPhone,
		Source:   source, // "ONBOARDING" or "ORDER"
	}

	if err := r.repo.Upsert(ctx, uc); err != nil {
		r.logger.Warn("failed to upsert refreshed user contact",
			baseLogger.UserID(userID),
			baseLogger.ErrorCode(err.Error()),
		)
	}

	return uc, nil
}

// UpsertFromKnownData stores mobile + email when we already have them
// (e.g., from the first eligibility request where caller sends mobile).
func (r *UserContactResolver) UpsertFromKnownData(ctx context.Context, userID, mobile, email, source string) error {
	uc := &entity.UserContact{
		UserID:   userID,
		Mobile:   mobile,
		Email:    email,
		RawPhone: "+91" + mobile,
		Source:   source,
	}
	return r.repo.Upsert(ctx, uc)
}
