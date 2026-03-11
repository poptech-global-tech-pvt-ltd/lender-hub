package response

import "time"

// CustomerStatusResponse is the canonical response for POST /v1/payin3/customer-status
type CustomerStatusResponse struct {
	UserID             string    `json:"userId"`
	Lender             string    `json:"lender"`
	PreApproved        bool      `json:"preApproved"`
	OnboardingRequired bool      `json:"onboardingRequired"`
	OnboardingDone     bool      `json:"onboardingDone"`
	NTBEligible        *bool     `json:"ntbEligible"`
	AvailableLimit     float64   `json:"availableLimit"`
	CheckedAt          time.Time `json:"checkedAt"`
}
