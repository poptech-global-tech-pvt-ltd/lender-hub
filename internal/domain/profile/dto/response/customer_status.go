package response

// CustomerStatusResponse is for POST /v1/payin3/customer-status
type CustomerStatusResponse struct {
	UserID               string  `json:"userId"`
	Provider             string  `json:"provider"`
	PreApproved          bool    `json:"preApproved"`
	OnboardingRequired   bool    `json:"onboardingRequired"`
	CustomerInfoRequired bool    `json:"customerInfoRequired"`
	AvailableLimit       float64 `json:"availableLimit"`
	NTBEligible          *bool   `json:"ntbEligible"`
}
