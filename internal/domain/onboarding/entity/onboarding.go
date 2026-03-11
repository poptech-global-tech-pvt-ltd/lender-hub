package entity

import "time"

// Onboarding represents an onboarding attempt
type Onboarding struct {
	ID                     int64
	OnboardingID           string
	ProviderOnboardingID   *string
	UserID                 string
	MerchantID             string
	Provider               string
	Mobile                 string
	Source                 string
	Channel                *string
	Status                 OnboardingStatus
	LastStep               *OnboardingStep
	RejectionReasonCode    *string
	RejectionReasonMessage *string
	COFEligible            bool
	RedirectURL            *string
	IsRetryable            bool
	RetryCount             int
	NextRetryAt            *time.Time
	LastRetryAt            *time.Time
	RawRequest             []byte
	RawResponse            []byte
	CreatedAt              time.Time
	UpdatedAt              time.Time
}
