package service

import (
	"context"
	"time"

	"lending-hub-service/internal/domain/profile/entity"
	"lending-hub-service/internal/domain/profile/port"
	sharedErrors "lending-hub-service/internal/shared/errors"
)

// ProfileUpdater handles profile state updates
type ProfileUpdater struct {
	repo      port.ProfileRepository
	publisher port.ProfileEventPublisher
}

// NewProfileUpdater creates a new ProfileUpdater
func NewProfileUpdater(repo port.ProfileRepository, publisher port.ProfileEventPublisher) *ProfileUpdater {
	return &ProfileUpdater{
		repo:      repo,
		publisher: publisher,
	}
}

// UpdateOnOnboardingSuccess updates profile when onboarding completes successfully
func (u *ProfileUpdater) UpdateOnOnboardingSuccess(ctx context.Context, userID, lender string, limit float64) error {
	profile, err := u.repo.GetForUpdate(ctx, userID, lender)
	if err != nil {
		return err
	}
	if profile == nil {
		return sharedErrors.New(sharedErrors.CodeUserNotFound, 404, "profile not found")
	}

	previousStatus := string(profile.Status)
	if !profile.CanTransitionTo(entity.ProfileActive) {
		return sharedErrors.New(sharedErrors.CodeInvalidTransition, 400, "cannot transition from "+previousStatus+" to ACTIVE")
	}

	// Update profile
	profile.Status = entity.ProfileActive
	profile.OnboardingDone = true
	profile.CreditLine.Limit = limit
	profile.CreditLine.AvailableLimit = limit
	profile.CreditLine.Currency = "INR"
	now := time.Now()
	profile.LastLimitRefresh = &now

	if err := u.repo.Upsert(ctx, profile); err != nil {
		return err
	}

	// Publish event
	u.publisher.Publish(ctx, &port.ProfileEvent{
		Type:           port.EventProfileActivated,
		UserID:         userID,
		Lender:         lender,
		PreviousStatus: previousStatus,
		NewStatus:      string(entity.ProfileActive),
		CreditLimit:    limit,
		AvailableLimit: limit,
		IsBlocked:      false,
	})

	return nil
}

// BlockUser blocks a user profile
func (u *ProfileUpdater) BlockUser(ctx context.Context, userID, lender, reason, source string) error {
	profile, err := u.repo.GetForUpdate(ctx, userID, lender)
	if err != nil {
		return err
	}
	if profile == nil {
		return sharedErrors.New(sharedErrors.CodeUserNotFound, 404, "profile not found")
	}

	previousStatus := string(profile.Status)
	profile.Block.IsBlocked = true
	profile.Block.Reason = reason
	profile.Block.Source = source
	profile.Block.NextEligibleAt = nil // Permanent block for now

	// Transition to BLOCKED if not already
	if profile.Status != entity.ProfileBlocked {
		if !profile.CanTransitionTo(entity.ProfileBlocked) {
			return sharedErrors.New(sharedErrors.CodeInvalidTransition, 400, "cannot transition to BLOCKED")
		}
		profile.Status = entity.ProfileBlocked
	}

	if err := u.repo.Upsert(ctx, profile); err != nil {
		return err
	}

	// Publish event
	u.publisher.Publish(ctx, &port.ProfileEvent{
		Type:           port.EventProfileBlocked,
		UserID:         userID,
		Lender:         lender,
		PreviousStatus: previousStatus,
		NewStatus:      string(entity.ProfileBlocked),
		CreditLimit:    profile.CreditLine.Limit,
		AvailableLimit: profile.CreditLine.AvailableLimit,
		IsBlocked:      true,
		BlockReason:    reason,
		BlockSource:    source,
	})

	return nil
}

// UnblockUser unblocks a user profile
func (u *ProfileUpdater) UnblockUser(ctx context.Context, userID, lender string) error {
	profile, err := u.repo.GetForUpdate(ctx, userID, lender)
	if err != nil {
		return err
	}
	if profile == nil {
		return sharedErrors.New(sharedErrors.CodeUserNotFound, 404, "profile not found")
	}

	if !profile.Block.IsBlocked {
		return nil // Already unblocked
	}

	previousStatus := string(profile.Status)
	profile.Block.IsBlocked = false
	profile.Block.Reason = ""
	profile.Block.Source = ""
	profile.Block.NextEligibleAt = nil

	// Transition to ACTIVE if was BLOCKED
	if profile.Status == entity.ProfileBlocked {
		if !profile.CanTransitionTo(entity.ProfileActive) {
			return sharedErrors.New(sharedErrors.CodeInvalidTransition, 400, "cannot transition to ACTIVE")
		}
		profile.Status = entity.ProfileActive
	}

	if err := u.repo.Upsert(ctx, profile); err != nil {
		return err
	}

	// Publish event
	u.publisher.Publish(ctx, &port.ProfileEvent{
		Type:           port.EventProfileUnblocked,
		UserID:         userID,
		Lender:         lender,
		PreviousStatus: previousStatus,
		NewStatus:      string(profile.Status),
		CreditLimit:    profile.CreditLine.Limit,
		AvailableLimit: profile.CreditLine.AvailableLimit,
		IsBlocked:      false,
	})

	return nil
}

// UpdateLimit updates the available credit limit
func (u *ProfileUpdater) UpdateLimit(ctx context.Context, userID, lender string, newAvailable float64) error {
	profile, err := u.repo.GetForUpdate(ctx, userID, lender)
	if err != nil {
		return err
	}
	if profile == nil {
		return sharedErrors.New(sharedErrors.CodeUserNotFound, 404, "profile not found")
	}

	profile.CreditLine.AvailableLimit = newAvailable
	now := time.Now()
	profile.LastLimitRefresh = &now

	if err := u.repo.Upsert(ctx, profile); err != nil {
		return err
	}

	// Publish event
	u.publisher.Publish(ctx, &port.ProfileEvent{
		Type:           port.EventLimitUpdated,
		UserID:         userID,
		Lender:         lender,
		PreviousStatus: string(profile.Status),
		NewStatus:      string(profile.Status),
		CreditLimit:    profile.CreditLine.Limit,
		AvailableLimit: newAvailable,
		IsBlocked:      profile.Block.IsBlocked,
	})

	return nil
}
