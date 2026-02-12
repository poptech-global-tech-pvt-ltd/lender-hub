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
	return &entity.UserProfile{
		ID:                 model.ID,
		UserID:             model.UserID,
		Lender:             model.Lender,
		CurrentStatus:      model.CurrentStatus,
		OnboardingDone:     model.OnboardingDone,
		NTBStatus:          model.NTBStatus,
		CreditLimit:        model.CreditLimit,
		AvailableLimit:     model.AvailableLimit,
		CreditLineActive:   model.CreditLineActive,
		CreditLineSummary:  model.CreditLineSummary,
		IsBlocked:          model.IsBlocked,
		BlockReason:        model.BlockReason,
		BlockSource:        model.BlockSource,
		NextEligibleAt:     model.NextEligibleAt,
		LastOnboardingID:   model.LastOnboardingID,
		LastLimitRefreshAt: model.LastLimitRefreshAt,
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
	}
}

// toModel converts domain entity to GORM model
func toModel(e *entity.UserProfile) *infra.LenderUser {
	return &infra.LenderUser{
		ID:                 e.ID,
		UserID:             e.UserID,
		Lender:             e.Lender,
		CurrentStatus:      e.CurrentStatus,
		OnboardingDone:     e.OnboardingDone,
		NTBStatus:          e.NTBStatus,
		CreditLimit:        e.CreditLimit,
		AvailableLimit:     e.AvailableLimit,
		CreditLineActive:   e.CreditLineActive,
		CreditLineSummary:  e.CreditLineSummary,
		IsBlocked:          e.IsBlocked,
		BlockReason:        e.BlockReason,
		BlockSource:        e.BlockSource,
		NextEligibleAt:     e.NextEligibleAt,
		LastOnboardingID:   e.LastOnboardingID,
		LastLimitRefreshAt: e.LastLimitRefreshAt,
		CreatedAt:          e.CreatedAt,
		UpdatedAt:          e.UpdatedAt,
	}
}
