package response

import "time"

// UserProfileResponse is the combined response for GET /v1/payin3/profile/:userId
// Status is NEVER empty: ACTIVE|INELIGIBLE|NOT_STARTED|BLOCKED
type UserProfileResponse struct {
	UserID  string `json:"userId"`
	Lender  string `json:"lender"`
	Status  string `json:"status"` // NEVER empty

	// From CustomerStatus
	PreApproved        bool  `json:"preApproved"`
	OnboardingRequired bool  `json:"onboardingRequired"`
	OnboardingDone     bool  `json:"onboardingDone"`
	NTBEligible        *bool `json:"ntbEligible"`

	// From Eligibility (only present when amount provided)
	TxnEligible       bool      `json:"txnEligible"`
	EligibilityCode   string    `json:"eligibilityCode,omitempty"`
	EligibilityReason string    `json:"eligibilityReason,omitempty"`
	AvailableLimit    float64   `json:"availableLimit"`
	CreditLimit       float64   `json:"creditLimit"`
	EmiPlans          []EmiPlan `json:"emiPlans,omitempty"`

	LastCheckedAt time.Time `json:"lastCheckedAt"`
}
