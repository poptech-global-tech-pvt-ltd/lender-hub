package response

// UserProfileResponse is for GET /v1/payin3/profile/:userId (combined)
type UserProfileResponse struct {
	UserID               string    `json:"userId"`
	Provider             string    `json:"provider"`
	PreApproved          bool      `json:"preApproved"`
	OnboardingRequired   bool      `json:"onboardingRequired"`
	CustomerInfoRequired bool      `json:"customerInfoRequired"`
	NTBEligible          *bool     `json:"ntbEligible"`
	TxnEligible          *bool     `json:"txnEligible,omitempty"`
	AvailableLimit       float64   `json:"availableLimit"`
	EmiPlans             []EmiPlan `json:"emiPlans,omitempty"`
	ExistingUser         *bool     `json:"existingUser,omitempty"`
	Status               string    `json:"status"`
	ReasonCode           string    `json:"reasonCode,omitempty"`
	ReasonMessage        string    `json:"reasonMessage,omitempty"`
}
