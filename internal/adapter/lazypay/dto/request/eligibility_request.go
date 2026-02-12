package request

import "lending-hub-service/internal/adapter/lazypay/dto/common"

// LPEligibilityRequest matches Lazypay /v7/payment/eligibility
type LPEligibilityRequest struct {
	AccessKey   string              `json:"accessKey"`
	MerchantID  string              `json:"merchantId"`
	User        common.LPUserDetails `json:"user"`
	OrderAmount common.LPAmount     `json:"orderAmount"`
	Signature   string              `json:"signature"`
	Source      string              `json:"source,omitempty"`
}
