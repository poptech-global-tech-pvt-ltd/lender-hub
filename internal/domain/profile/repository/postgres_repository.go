package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	infra "lending-hub-service/internal/infrastructure/postgres"
	"lending-hub-service/internal/domain/profile/entity"
	"lending-hub-service/internal/domain/profile/port"
)

// postgresProfileRepository implements ProfileRepository using GORM
type postgresProfileRepository struct {
	db *gorm.DB
}

// NewProfileRepository creates a new Postgres-backed ProfileRepository
func NewProfileRepository(db *gorm.DB) port.ProfileRepository {
	return &postgresProfileRepository{db: db}
}

// Get returns a user profile by userId+lender, or nil if not found
func (r *postgresProfileRepository) Get(ctx context.Context, userID, lender string) (*entity.UserProfile, error) {
	var model infra.LenderUser
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND lender = ?", userID, lender).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toEntity(&model), nil
}

// GetForUpdate returns a user profile row locked for update (FOR UPDATE)
func (r *postgresProfileRepository) GetForUpdate(ctx context.Context, userID, lender string) (*entity.UserProfile, error) {
	var model infra.LenderUser
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND lender = ?", userID, lender).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toEntity(&model), nil
}

// Upsert creates or updates a profile row
func (r *postgresProfileRepository) Upsert(ctx context.Context, profile *entity.UserProfile) error {
	model := toModel(profile)
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "lender"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"current_status", "onboarding_done", "ntb_status",
				"credit_limit", "available_limit", "credit_line_active",
				"credit_line_summary", "is_blocked", "block_reason",
				"block_source", "next_eligible_at", "last_onboarding_id",
				"last_limit_refresh_at", "updated_at",
			}),
		}).
		Create(&model).Error
}

// toEntity converts GORM model to domain entity
func toEntity(model *infra.LenderUser) *entity.UserProfile {
	profile := &entity.UserProfile{
		UserID:            model.UserID,
		Lender:            model.Lender,
		Status:            entity.ProfileStatus(model.CurrentStatus),
		OnboardingDone:    model.OnboardingDone != nil && *model.OnboardingDone,
		CreditLineSummary: model.CreditLineSummary,
		LastOnboardingID:  model.LastOnboardingID,
		LastLimitRefresh:  model.LastLimitRefreshAt,
		CreatedAt:         model.CreatedAt,
		UpdatedAt:         model.UpdatedAt,
	}

	// Map credit line
	if model.CreditLimit != nil {
		profile.CreditLine.Limit = *model.CreditLimit
	}
	if model.AvailableLimit != nil {
		profile.CreditLine.AvailableLimit = *model.AvailableLimit
	}
	profile.CreditLine.Currency = "INR"

	// Map block info
	if model.IsBlocked != nil {
		profile.Block.IsBlocked = *model.IsBlocked
	}
	if model.BlockReason != nil {
		profile.Block.Reason = *model.BlockReason
	}
	if model.BlockSource != nil {
		profile.Block.Source = *model.BlockSource
	}
	profile.Block.NextEligibleAt = model.NextEligibleAt

	return profile
}

// toModel converts domain entity to GORM model
func toModel(e *entity.UserProfile) *infra.LenderUser {
	onboardingDone := e.OnboardingDone
	isBlocked := e.Block.IsBlocked
	creditLimit := e.CreditLine.Limit
	availableLimit := e.CreditLine.AvailableLimit

	var blockReason *string
	if e.Block.Reason != "" {
		blockReason = &e.Block.Reason
	}

	var blockSource *string
	if e.Block.Source != "" {
		blockSource = &e.Block.Source
	}

	return &infra.LenderUser{
		UserID:             e.UserID,
		Lender:             e.Lender,
		CurrentStatus:      string(e.Status),
		OnboardingDone:     &onboardingDone,
		CreditLimit:        &creditLimit,
		AvailableLimit:     &availableLimit,
		CreditLineActive:   e.CreditLine.Limit > 0,
		CreditLineSummary:  e.CreditLineSummary,
		IsBlocked:          &isBlocked,
		BlockReason:        blockReason,
		BlockSource:        blockSource,
		NextEligibleAt:     e.Block.NextEligibleAt,
		LastOnboardingID:   e.LastOnboardingID,
		LastLimitRefreshAt: e.LastLimitRefresh,
		CreatedAt:          e.CreatedAt,
		UpdatedAt:          e.UpdatedAt,
	}
}
