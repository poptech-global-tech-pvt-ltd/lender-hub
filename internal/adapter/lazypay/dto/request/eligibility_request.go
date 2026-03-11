package request

import "lending-hub-service/internal/adapter/lazypay/dto/common"

// LPEligibilityRequest matches Lazypay POST /api/lazypay/v7/payment/eligibility
type LPEligibilityRequest struct {
	UserDetails common.LPUserDetails `json:"userDetails"`
	Amount      common.LPAmount      `json:"amount"`
	Source      string               `json:"source"`
}
