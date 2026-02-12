package postgres

import "time"

// LenderCustomerLink maps to lender_customer_link
type LenderCustomerLink struct {
	ID                 int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UserID             string    `gorm:"column:user_id;type:text;not null"`
	MerchantID         string    `gorm:"column:merchant_id;type:text;not null"`
	Provider           string    `gorm:"column:provider;type:text;not null"`
	Mobile             string    `gorm:"column:mobile;type:text;not null"`
	LatestOnboardingID *int64    `gorm:"column:latest_onboarding_id"`
	CreatedAt          time.Time `gorm:"column:created_at;not null"`
	UpdatedAt          time.Time `gorm:"column:updated_at;not null"`
}

func (LenderCustomerLink) TableName() string {
	return "lender_customer_link"
}

// LenderOnboarding maps to lender_onboarding
type LenderOnboarding struct {
	ID                     int64      `gorm:"column:id;primaryKey;autoIncrement"`
	OnboardingID           string     `gorm:"column:onboarding_id;type:text;not null"`
	ProviderOnboardingID   *string    `gorm:"column:provider_onboarding_id;type:text"`
	UserID                 string     `gorm:"column:user_id;type:text;not null"`
	MerchantID             string     `gorm:"column:merchant_id;type:text;not null"`
	Provider               string     `gorm:"column:provider;type:text;not null"`
	Mobile                 string     `gorm:"column:mobile;type:text;not null"`
	Source                 string     `gorm:"column:source;type:request_source;not null"`
	Channel                *string    `gorm:"column:channel;type:channel_type"`
	Status                 string     `gorm:"column:status;type:lender_onboarding_status;not null"`
	LastStep               *string    `gorm:"column:last_step;type:onboarding_step"`
	RejectionReasonCode    *string    `gorm:"column:rejection_reason_code;type:text"`
	RejectionReasonMessage *string    `gorm:"column:rejection_reason_message;type:text"`
	COFEligible            *bool       `gorm:"column:cof_eligible"`
	RedirectURL            *string    `gorm:"column:redirect_url;type:text"`
	IsRetryable            bool       `gorm:"column:is_retryable;not null"`
	RetryCount             int        `gorm:"column:retry_count;not null"`
	NextRetryAt            *time.Time `gorm:"column:next_retry_at"`
	LastRetryAt            *time.Time `gorm:"column:last_retry_at"`
	RawRequest             []byte     `gorm:"column:raw_request;type:jsonb"`
	RawResponse            []byte     `gorm:"column:raw_response;type:jsonb"`
	CreatedAt              time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt              time.Time  `gorm:"column:updated_at;not null"`
}

func (LenderOnboarding) TableName() string {
	return "lender_onboarding"
}

// LenderOnboardingEvent maps to lender_onboarding_events
type LenderOnboardingEvent struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Provider     string    `gorm:"column:provider;type:text;not null"`
	MerchantID   string    `gorm:"column:merchant_id;type:text;not null"`
	UserID       string    `gorm:"column:user_id;type:text;not null"`
	Mobile       string    `gorm:"column:mobile;type:text;not null"`
	OnboardingID string    `gorm:"column:onboarding_id;type:text;not null"`
	EventType    string    `gorm:"column:event_type;type:text;not null"`
	Status       string    `gorm:"column:status;type:text;not null"`
	Step         *string   `gorm:"column:step;type:onboarding_step"`
	ErrorCode    *string   `gorm:"column:error_code;type:text"`
	Message      *string   `gorm:"column:message;type:text"`
	EventTime    time.Time `gorm:"column:event_time;not null"`
	RawPayload   []byte    `gorm:"column:raw_payload;type:jsonb"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
}

func (LenderOnboardingEvent) TableName() string {
	return "lender_onboarding_events"
}
