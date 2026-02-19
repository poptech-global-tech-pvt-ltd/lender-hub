package mapper

import (
	lpCommon "lending-hub-service/internal/adapter/lazypay/dto/common"
	lpReq "lending-hub-service/internal/adapter/lazypay/dto/request"
	lpResp "lending-hub-service/internal/adapter/lazypay/dto/response"
	onbResp "lending-hub-service/internal/domain/onboarding/dto/response"
)

// OnboardingMapper maps between canonical and Lazypay onboarding formats
type OnboardingMapper struct {
	cfg *OnboardingMapperConfig
}

// OnboardingMapperConfig holds config needed for mapping
type OnboardingMapperConfig struct {
	SubMerchantID string
	ReturnURL     string
}

// NewOnboardingMapper creates a new onboarding mapper
func NewOnboardingMapper(cfg *OnboardingMapperConfig) *OnboardingMapper {
	return &OnboardingMapper{cfg: cfg}
}

// ToLPRequest converts canonical onboarding input → LP format
// Postman contract: { customParams: { subMerchantId }, userDetails: { mobile, email }, returnUrl, source }
func (m *OnboardingMapper) ToLPRequest(mobile, email string) *lpReq.LPOnboardingRequest {
	return &lpReq.LPOnboardingRequest{
		CustomParams: lpCommon.LPCustomParams{
			SubMerchantID: m.cfg.SubMerchantID,
		},
		UserDetails: lpCommon.NewLPUserDetails(mobile, email),
		ReturnURL:   m.cfg.ReturnURL,
		Source:      "website",
	}
}

// FromLPOnboardingResponse converts LP response → canonical OnboardingResponse
func FromLPOnboardingResponse(
	lp *lpResp.LPOnboardingResponse,
	onboardingTxnID string,
) *onbResp.OnboardingResponse {
	return &onbResp.OnboardingResponse{
		OnboardingID:    lp.OnboardingID,
		OnboardingTxnID: onboardingTxnID,
		Provider:        "LAZYPAY",
		RedirectURL:     lp.RedirectURL,
		Status:          lp.Status,
	}
}
