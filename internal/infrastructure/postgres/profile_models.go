package postgres

import "time"

// LenderUser maps to table lender_user
type LenderUser struct {
	ID                   int64      `gorm:"column:id;primaryKey;autoIncrement"`
	UserID               string     `gorm:"column:user_id;type:text;not null"`
	Lender               string     `gorm:"column:lender;type:text;not null"`
	CurrentStatus        string     `gorm:"column:current_status;type:lender_profile_status;not null"`
	OnboardingDone       *bool      `gorm:"column:onboarding_done"`
	OnboardingRequired   *bool      `gorm:"column:onboarding_required"`
	NTBStatus            *bool      `gorm:"column:ntb_status"`
	PreApproved          *bool      `gorm:"column:pre_approved"`
	CustomerInfoRequired *bool      `gorm:"column:customer_info_required"`
	NTBEligible          *bool      `gorm:"column:ntb_eligible"`
	TxnEligible          *bool      `gorm:"column:txn_eligible"`
	ExistingUser         *bool      `gorm:"column:existing_user"`
	EligibilityRespID    *string    `gorm:"column:eligibility_resp_id;type:text"`
	CreditLimit          *float64   `gorm:"column:credit_limit;type:numeric(14,2)"`
	AvailableLimit       *float64   `gorm:"column:available_limit;type:numeric(14,2)"`
	CreditLineActive     bool       `gorm:"column:credit_line_active;not null"`
	CreditLineSummary    []byte     `gorm:"column:credit_line_summary;type:jsonb"`
	IsBlocked            *bool      `gorm:"column:is_blocked"`
	BlockReason          *string    `gorm:"column:block_reason;type:text"`
	BlockSource          *string    `gorm:"column:block_source;type:text"`
	NextEligibleAt       *time.Time `gorm:"column:next_eligible_at"`
	LastOnboardingID     *int64     `gorm:"column:last_onboarding_id"`
	LastLimitRefreshAt   *time.Time `gorm:"column:last_limit_refresh_at"`
	LastEligibilityAt    *time.Time `gorm:"column:last_eligibility_at"`
	LastStatusCheckAt    *time.Time `gorm:"column:last_status_check_at"`
	CreatedAt            time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt            time.Time  `gorm:"column:updated_at;not null"`
}

func (LenderUser) TableName() string {
	return "lender_user"
}
