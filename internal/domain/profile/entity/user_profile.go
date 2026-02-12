package entity

import "time"

// UserProfile represents a user's profile with a lender
type UserProfile struct {
	ID                 int64
	UserID             string
	Lender             string
	CurrentStatus      string
	OnboardingDone     *bool
	NTBStatus          *bool
	CreditLimit        *float64
	AvailableLimit     *float64
	CreditLineActive   bool
	CreditLineSummary  []byte
	IsBlocked          *bool
	BlockReason        *string
	BlockSource        *string
	NextEligibleAt     *time.Time
	LastOnboardingID   *int64
	LastLimitRefreshAt *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
