package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	infra "lending-hub-service/internal/infrastructure/postgres"
	"lending-hub-service/internal/domain/onboarding/entity"
	"lending-hub-service/internal/domain/onboarding/port"
)

// postgresOnboardingRepository implements OnboardingRepository using GORM
type postgresOnboardingRepository struct {
	db *gorm.DB
}

// NewOnboardingRepository creates a new Postgres-backed OnboardingRepository
func NewOnboardingRepository(db *gorm.DB) port.OnboardingRepository {
	return &postgresOnboardingRepository{db: db}
}

// Create creates a new onboarding record
func (r *postgresOnboardingRepository) Create(ctx context.Context, ob *entity.Onboarding) error {
	model := toModel(ob)
	return r.db.WithContext(ctx).Create(&model).Error
}

// GetByOnboardingID retrieves an onboarding by onboarding_id
func (r *postgresOnboardingRepository) GetByOnboardingID(ctx context.Context, onboardingID string) (*entity.Onboarding, error) {
	var model infra.LenderOnboarding
	err := r.db.WithContext(ctx).
		Where("onboarding_id = ?", onboardingID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toEntity(&model), nil
}

// GetForUpdate retrieves an onboarding with row lock for update
func (r *postgresOnboardingRepository) GetForUpdate(ctx context.Context, onboardingID string) (*entity.Onboarding, error) {
	var model infra.LenderOnboarding
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("onboarding_id = ?", onboardingID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toEntity(&model), nil
}

// Update updates an existing onboarding record
func (r *postgresOnboardingRepository) Update(ctx context.Context, ob *entity.Onboarding) error {
	model := toModel(ob)
	return r.db.WithContext(ctx).
		Where("onboarding_id = ?", ob.OnboardingID).
		Updates(&model).Error
}

// toEntity converts GORM model to domain entity
func toEntity(model *infra.LenderOnboarding) *entity.Onboarding {
	return &entity.Onboarding{
		ID:                     model.ID,
		OnboardingID:           model.OnboardingID,
		ProviderOnboardingID:   model.ProviderOnboardingID,
		UserID:                 model.UserID,
		MerchantID:             model.MerchantID,
		Provider:               model.Provider,
		Mobile:                 model.Mobile,
		Source:                 model.Source,
		Channel:                model.Channel,
		Status:                 model.Status,
		LastStep:               model.LastStep,
		RejectionReasonCode:    model.RejectionReasonCode,
		RejectionReasonMessage: model.RejectionReasonMessage,
		COFEligible:            model.COFEligible,
		RedirectURL:            model.RedirectURL,
		IsRetryable:            model.IsRetryable,
		RetryCount:             model.RetryCount,
		NextRetryAt:            model.NextRetryAt,
		LastRetryAt:            model.LastRetryAt,
		RawRequest:             model.RawRequest,
		RawResponse:            model.RawResponse,
		CreatedAt:              model.CreatedAt,
		UpdatedAt:              model.UpdatedAt,
	}
}

// toModel converts domain entity to GORM model
func toModel(e *entity.Onboarding) *infra.LenderOnboarding {
	return &infra.LenderOnboarding{
		ID:                     e.ID,
		OnboardingID:           e.OnboardingID,
		ProviderOnboardingID:   e.ProviderOnboardingID,
		UserID:                 e.UserID,
		MerchantID:             e.MerchantID,
		Provider:               e.Provider,
		Mobile:                 e.Mobile,
		Source:                 e.Source,
		Channel:                e.Channel,
		Status:                 e.Status,
		LastStep:               e.LastStep,
		RejectionReasonCode:    e.RejectionReasonCode,
		RejectionReasonMessage: e.RejectionReasonMessage,
		COFEligible:            e.COFEligible,
		RedirectURL:            e.RedirectURL,
		IsRetryable:            e.IsRetryable,
		RetryCount:             e.RetryCount,
		NextRetryAt:            e.NextRetryAt,
		LastRetryAt:            e.LastRetryAt,
		RawRequest:             e.RawRequest,
		RawResponse:            e.RawResponse,
		CreatedAt:              e.CreatedAt,
		UpdatedAt:              e.UpdatedAt,
	}
}
