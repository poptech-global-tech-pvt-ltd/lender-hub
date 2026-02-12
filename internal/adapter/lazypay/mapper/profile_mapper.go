package mapper

import (
	"fmt"

	profileReq "lending-hub-service/internal/domain/profile/dto/request"
	profileResp "lending-hub-service/internal/domain/profile/dto/response"
	lpCommon "lending-hub-service/internal/adapter/lazypay/dto/common"
	lpReq "lending-hub-service/internal/adapter/lazypay/dto/request"
	lpResp "lending-hub-service/internal/adapter/lazypay/dto/response"
)

// ToLPEligibilityRequest converts canonical CustomerStatusRequest → LP format
func ToLPEligibilityRequest(
	req profileReq.CustomerStatusRequest,
	accessKey, merchantID string,
	signature string,
) *lpReq.LPEligibilityRequest {
	// Calculate order amount from context (default to 0 if not provided)
	orderAmount := 0.0
	if req.Context.Platform != "" {
		// In real implementation, might extract from context
		// For now, use a default
	}

	return &lpReq.LPEligibilityRequest{
		AccessKey:   accessKey,
		MerchantID:  merchantID,
		User: lpCommon.LPUserDetails{
			Mobile: req.Mobile,
			Email:  req.Email,
		},
		OrderAmount: lpCommon.LPAmount{
			Value:    fmt.Sprintf("%.2f", orderAmount),
			Currency: "INR",
		},
		Signature: signature,
		Source:    req.Source,
	}
}

// FromLPEligibilityResponse converts LP response → canonical CustomerStatusResponse
func FromLPEligibilityResponse(
	lp *lpResp.LPEligibilityResponse,
	userID string,
) *profileResp.CustomerStatusResponse {
	// Map LP status → canonical status
	var status profileResp.PayIn3Status
	if lp.Blocked {
		status = profileResp.StatusBlocked
	} else if lp.Status == "APPROVED" {
		if lp.CreditLineActive {
			status = profileResp.StatusActive
		} else {
			status = profileResp.StatusInProgress
		}
	} else if lp.Status == "REJECTED" {
		status = profileResp.StatusIneligible
	} else {
		status = profileResp.StatusNotStarted
	}

	// Map LP EMI plans → canonical EMI plans
	emiPlans := make([]profileResp.EmiPlan, len(lp.EMIPlans))
	for i, lpPlan := range lp.EMIPlans {
		emiPlans[i] = profileResp.EmiPlan{
			TenureMonths: lpPlan.Tenure,
			EmiAmount:    lpPlan.EMI,
			TotalAmount:  lpPlan.TotalAmount,
		}
	}

	return &profileResp.CustomerStatusResponse{
		UserID:              userID,
		Provider:            "LAZYPAY",
		PreApproved:         lp.PreApproved,
		AvailableLimit:      lp.AvailableLimit,
		CreditLineActive:    lp.CreditLineActive,
		OnboardingRequired:  !lp.CreditLineActive,
		Status:              status,
		ReasonCode:          lp.ReasonCode,
		ReasonMessage:       lp.ReasonMessage,
		EmiPlans:            emiPlans,
	}
}
