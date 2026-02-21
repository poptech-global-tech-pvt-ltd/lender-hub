package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	infra "lending-hub-service/internal/infrastructure/postgres"
	res "lending-hub-service/internal/domain/profile/dto/response"
	"lending-hub-service/internal/domain/profile/entity"
	"lending-hub-service/internal/domain/profile/port"
	lenderPkg "lending-hub-service/pkg/lender"
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


// UpsertFromEligibility updates lender_user from Eligibility API response
func (r *postgresProfileRepository) UpsertFromEligibility(ctx context.Context, userID, lender string, resp *res.EligibilityResponse) error {
	if lender == "" {
		lender = lenderPkg.Lazypay.String()
	}
	now := time.Now().UTC()
	status := mapStatusForDB(resp.ReasonCode, resp.TxnEligible)

	updates := map[string]interface{}{
		"txn_eligible":        resp.TxnEligible,
		"available_limit":     resp.AvailableLimit,
		"credit_limit":        resp.AvailableLimit,
		"existing_user":       resp.ExistingUser,
		"last_eligibility_at": now,
		"current_status":      status,
		"updated_at":          now,
	}
	if resp.EligibilityRespID != "" {
		updates["eligibility_resp_id"] = resp.EligibilityRespID
	}

	return r.upsertLenderUser(ctx, userID, lender, updates)
}

// UpsertFromCustomerStatus updates lender_user from Customer Status API response
func (r *postgresProfileRepository) UpsertFromCustomerStatus(ctx context.Context, userID, lender string, resp *res.CustomerStatusResponse) error {
	if lender == "" {
		lender = lenderPkg.Lazypay.String()
	}
	now := time.Now().UTC()
	status := deriveStatusFromCustomerStatus(resp)

	updates := map[string]interface{}{
		"pre_approved":           resp.PreApproved,
		"onboarding_required":    resp.OnboardingRequired,
		"customer_info_required": resp.CustomerInfoRequired,
		"ntb_eligible":           resp.NTBEligible,
		"available_limit":        resp.AvailableLimit,
		"credit_limit":           resp.AvailableLimit,
		"last_status_check_at":   now,
		"current_status":         status,
		"updated_at":             now,
	}
	onboardingDone := !resp.OnboardingRequired
	updates["onboarding_done"] = onboardingDone
	if resp.NTBEligible != nil {
		updates["ntb_status"] = *resp.NTBEligible
	}

	return r.upsertLenderUser(ctx, userID, lender, updates)
}

// UpsertFromCombined updates lender_user from combined UserProfileResponse
func (r *postgresProfileRepository) UpsertFromCombined(ctx context.Context, userID, lender string, profile *res.UserProfileResponse) error {
	if lender == "" {
		lender = lenderPkg.Lazypay.String()
	}
	now := time.Now().UTC()

	updates := map[string]interface{}{
		"pre_approved":           profile.PreApproved,
		"onboarding_required":    profile.OnboardingRequired,
		"customer_info_required": profile.CustomerInfoRequired,
		"ntb_eligible":           profile.NTBEligible,
		"available_limit":        profile.AvailableLimit,
		"credit_limit":           profile.AvailableLimit,
		"current_status":         profile.Status,
		"updated_at":             now,
	}
	updates["onboarding_done"] = !profile.OnboardingRequired
	if profile.NTBEligible != nil {
		updates["ntb_status"] = *profile.NTBEligible
	}
	if profile.TxnEligible != nil {
		updates["txn_eligible"] = *profile.TxnEligible
	}
	if profile.ExistingUser != nil {
		updates["existing_user"] = *profile.ExistingUser
	}

	return r.upsertLenderUser(ctx, userID, lender, updates)
}

func (r *postgresProfileRepository) upsertLenderUser(ctx context.Context, userID, lender string, updates map[string]interface{}) error {
	var existing infra.LenderUser
	err := r.db.WithContext(ctx).Where("user_id = ? AND lender = ?", userID, lender).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Insert new row
		model := infra.LenderUser{
			UserID:        userID,
			Lender:        lender,
			CurrentStatus: "NOT_STARTED",
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}
		if s, ok := updates["current_status"].(string); ok {
			model.CurrentStatus = s
		}
		if v, ok := updates["available_limit"].(float64); ok {
			model.AvailableLimit = &v
			model.CreditLimit = &v
		}
		if v, ok := updates["onboarding_done"].(bool); ok {
			model.OnboardingDone = &v
		}
		if v, ok := updates["pre_approved"].(bool); ok {
			model.PreApproved = &v
		}
		if v, ok := updates["onboarding_required"].(bool); ok {
			model.OnboardingRequired = &v
		}
		if v, ok := updates["customer_info_required"].(bool); ok {
			model.CustomerInfoRequired = &v
		}
		if v, ok := updates["ntb_eligible"].(*bool); ok && v != nil {
			model.NTBEligible = v
			model.NTBStatus = v
		}
		if v, ok := updates["txn_eligible"].(bool); ok {
			model.TxnEligible = &v
		}
		if v, ok := updates["existing_user"].(bool); ok {
			model.ExistingUser = &v
		}
		if v, ok := updates["eligibility_resp_id"].(string); ok && v != "" {
			model.EligibilityRespID = &v
		}
		if v, ok := updates["last_eligibility_at"].(time.Time); ok {
			model.LastEligibilityAt = &v
		}
		if v, ok := updates["last_status_check_at"].(time.Time); ok {
			model.LastStatusCheckAt = &v
		}
		return r.db.WithContext(ctx).Create(&model).Error
	}

	// Update existing
	return r.db.WithContext(ctx).Model(&existing).Updates(updates).Error
}

func mapStatusForDB(reasonCode string, txnEligible bool) string {
	if !txnEligible {
		if reasonCode == "COF_NOT_AVAILABLE" || reasonCode != "" {
			return "INELIGIBLE"
		}
		return "INELIGIBLE"
	}
	return "ACTIVE"
}

func deriveStatusFromCustomerStatus(cs *res.CustomerStatusResponse) string {
	if cs.OnboardingRequired {
		if cs.NTBEligible != nil && *cs.NTBEligible {
			return "NOT_STARTED" // NTB
		}
		return "NOT_STARTED"
	}
	if cs.PreApproved && cs.AvailableLimit > 0 {
		return "ACTIVE"
	}
	return "INELIGIBLE"
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
