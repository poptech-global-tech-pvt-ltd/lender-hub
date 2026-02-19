package response

// LPCustomerStatusResponse is for GET /api/lazypay/cof/v0/customer-status
type LPCustomerStatusResponse struct {
	CustomerInfoRequired bool   `json:"customerInfoRequired"`
	PreApprovalStatus    bool   `json:"preApprovalStatus"`
	OnboardingRequired   bool   `json:"onboardingRequired"`
	AvailableLimit       float64 `json:"availableLimit"`
	NTBEligible          *bool   `json:"ntbEligible"`
}
