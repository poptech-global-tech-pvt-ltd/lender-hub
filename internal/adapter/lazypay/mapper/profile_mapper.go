package mapper

import (
	lpCommon "lending-hub-service/internal/adapter/lazypay/dto/common"
	lpReq "lending-hub-service/internal/adapter/lazypay/dto/request"
	lpResp "lending-hub-service/internal/adapter/lazypay/dto/response"
	"lending-hub-service/internal/domain/profile/port"
)

// ProfileMapper maps between Lazypay and domain port types
type ProfileMapper struct{}

// NewProfileMapper creates a new profile mapper
func NewProfileMapper() *ProfileMapper {
	return &ProfileMapper{}
}

// ToLPEligibilityRequest converts canonical input → LP format
func (m *ProfileMapper) ToLPEligibilityRequest(mobile, email string, amount float64) *lpReq.LPEligibilityRequest {
	return &lpReq.LPEligibilityRequest{
		UserDetails: lpCommon.NewLPUserDetails(mobile, email),
		Amount: lpCommon.LPAmount{
			Value:    lpCommon.FormatAmount(amount),
			Currency: "INR",
		},
		Source: "website",
	}
}

// ToLPCustomerStatusRequest converts canonical input → LP format
func (m *ProfileMapper) ToLPCustomerStatusRequest(mobile, email string) *lpReq.LPCustomerStatusBody {
	return &lpReq.LPCustomerStatusBody{
		UserDetails: lpCommon.NewLPUserDetails(mobile, email),
	}
}

// MapEligibilityResponse maps LP Eligibility response → port.EligibilityResult
// Uses COF section only (ignores BNPL)
func MapEligibilityResponse(lp *lpResp.LPEligibilityResponse) *port.EligibilityResult {
	if lp.COF == nil {
		return &port.EligibilityResult{
			TxnEligible:       false,
			EligibilityCode:   "COF_NOT_AVAILABLE",
			EligibilityReason: "COF section not available",
			EligibilityResponseID: lp.EligibilityResponseID,
			ExistingUser:      lp.ExistingUser,
		}
	}
	cof := lp.COF
	return &port.EligibilityResult{
		TxnEligible:           cof.TxnEligibility,
		EligibilityCode:       cof.Code,
		EligibilityReason:     cof.Reason,
		AvailableLimit:        cof.AvailableLimit,
		EmiPlans:              mapEmiPlansToResult(cof.EmiPlans),
		EligibilityResponseID: lp.EligibilityResponseID,
		ExistingUser:          lp.ExistingUser,
	}
}

// MapCustomerStatusResponse maps LP Customer Status response → port.CustomerStatusResult
func MapCustomerStatusResponse(lp *lpResp.LPCustomerStatusResponse) *port.CustomerStatusResult {
	return &port.CustomerStatusResult{
		PreApproved:        lp.PreApprovalStatus,
		OnboardingRequired: lp.OnboardingRequired,
		NTBEligible:        lp.NTBEligible,
		AvailableLimit:     lp.AvailableLimit,
	}
}

func mapEmiPlansToResult(lpPlans []lpResp.LPEmiPlan) []port.EmiPlanResult {
	if len(lpPlans) == 0 {
		return nil
	}
	plans := make([]port.EmiPlanResult, len(lpPlans))
	for i, p := range lpPlans {
		plans[i] = port.EmiPlanResult{
			Tenure:             p.Tenure,
			EMI:                p.Emi,
			InterestRate:       p.InterestRate,
			Principal:          p.Principal,
			TotalPayableAmount: p.TotalPayableAmount,
			FirstEmiDueDate:    p.FirstEmiDueDate,
			Type:               p.Type,
		}
	}
	return plans
}
