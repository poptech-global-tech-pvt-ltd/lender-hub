package request

import "lending-hub-service/internal/adapter/lazypay/dto/common"

// LPCustomerStatusBody matches Lazypay GET /api/lazypay/cof/v0/customer-status body
type LPCustomerStatusBody struct {
	UserDetails common.LPUserDetails `json:"userDetails"`
}
