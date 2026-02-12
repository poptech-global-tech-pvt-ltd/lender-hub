package entity

import "time"

// UserProfile is the aggregate root for a user's Pay-in-3 profile
type UserProfile struct {
	UserID            string
	Lender            string
	Status            ProfileStatus
	OnboardingDone    bool
	CreditLine        CreditLine
	Block             BlockInfo
	CreditLineSummary []byte     // JSONB: cached eligibility data
	LastOnboardingID  *int64
	LastLimitRefresh  *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// CanTransitionTo delegates to ProfileStatus
func (p *UserProfile) CanTransitionTo(next ProfileStatus) bool {
	return p.Status.CanTransitionTo(next)
}

// IsEligible returns true if user can currently transact
func (p *UserProfile) IsEligible() bool {
	return p.Status == ProfileActive && !p.Block.IsBlocked && p.CreditLine.AvailableLimit > 0
}
