package response

type OnboardingResponse struct {
	OnboardingID    string `json:"onboardingId"`
	OnboardingTxnID string `json:"onboardingTxnId"`
	Provider        string `json:"provider"`
	RedirectURL     string `json:"redirectUrl"`
	Status          string `json:"status"`
}
