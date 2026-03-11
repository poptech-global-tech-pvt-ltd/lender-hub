package response

// LPOnboardingResponse matches Lazypay onboarding creation response
type LPOnboardingResponse struct {
	OnboardingID string `json:"onboardingId"`
	Status       string `json:"status"` // "PENDING", "SUCCESS", etc.
	RedirectURL  string `json:"redirectUrl"`
	COFEligible  bool   `json:"cofEligible"`
}
