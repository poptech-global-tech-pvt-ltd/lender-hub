package request

import "lending-hub-service/internal/adapter/lazypay/dto/common"

// LPOnboardingRequest matches Lazypay POST /api/lazypay/cof/v0/standalone/initiate-onboarding
type LPOnboardingRequest struct {
	CustomParams common.LPCustomParams `json:"customParams"`
	UserDetails  common.LPUserDetails  `json:"userDetails"`
	ReturnURL    string                `json:"returnUrl"`
	Source       string                `json:"source"`
}
