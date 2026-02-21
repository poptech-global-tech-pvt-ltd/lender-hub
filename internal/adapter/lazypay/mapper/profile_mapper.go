package mapper

import (
	lpCommon "lending-hub-service/internal/adapter/lazypay/dto/common"
	lpReq "lending-hub-service/internal/adapter/lazypay/dto/request"
	lpResp "lending-hub-service/internal/adapter/lazypay/dto/response"
	profileResp "lending-hub-service/internal/domain/profile/dto/response"
	"lending-hub-service/pkg/lender"
)

// ProfileMapper maps between canonical and Lazypay profile formats
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

// FromLPEligibilityResponse maps LP Eligibility response → canonical EligibilityResponse
// Filters on COF section only (ignore BNPL)
func FromLPEligibilityResponse(lp *lpResp.LPEligibilityResponse, userID string) *profileResp.EligibilityResponse {
	if lp.COF == nil {
		return &profileResp.EligibilityResponse{
			UserID:      userID,
			Provider:    lender.Lazypay.String(),
			TxnEligible: false,
			ReasonCode:  "COF_NOT_AVAILABLE",
		}
	}
	cof := lp.COF
	emiPlans := mapEmiPlans(cof.EmiPlans)
	return &profileResp.EligibilityResponse{
		UserID:            userID,
		Provider:          lender.Lazypay.String(),
		TxnEligible:       cof.TxnEligibility,
		AvailableLimit:    cof.AvailableLimit,
		EmiPlans:          emiPlans,
		ExistingUser:      lp.ExistingUser,
		ReasonCode:        cof.Code,
		ReasonMessage:     cof.Reason,
		EligibilityRespID: lp.EligibilityResponseID,
	}
}

func mapEmiPlans(lpPlans []lpResp.LPEmiPlan) []profileResp.EmiPlan {
	plans := make([]profileResp.EmiPlan, len(lpPlans))
	for i, p := range lpPlans {
		plans[i] = profileResp.EmiPlan{
			Tenure:             p.Tenure,
			Emi:                p.Emi,
			InterestRate:       p.InterestRate,
			Principal:          p.Principal,
			TotalPayableAmount: p.TotalPayableAmount,
			FirstEmiDueDate:    p.FirstEmiDueDate,
			Type:               p.Type,
		}
	}
	return plans
}

// FromLPCustomerStatusResponse maps LP Customer Status response → canonical CustomerStatusResponse
func FromLPCustomerStatusResponse(lp *lpResp.LPCustomerStatusResponse, userID string) *profileResp.CustomerStatusResponse {
	return &profileResp.CustomerStatusResponse{
		UserID:               userID,
		Provider:             lender.Lazypay.String(),
		PreApproved:          lp.PreApprovalStatus,
		OnboardingRequired:   lp.OnboardingRequired,
		CustomerInfoRequired: lp.CustomerInfoRequired,
		AvailableLimit:       lp.AvailableLimit,
		NTBEligible:          lp.NTBEligible,
	}
}

// MergeToUserProfile merges CustomerStatus + Eligibility into combined UserProfileResponse
func MergeToUserProfile(
	cs *profileResp.CustomerStatusResponse,
	elig *profileResp.EligibilityResponse,
	userID string,
) *profileResp.UserProfileResponse {
	resp := &profileResp.UserProfileResponse{
		UserID:               userID,
		Provider:             lender.Lazypay.String(),
		PreApproved:          cs.PreApproved,
		OnboardingRequired:   cs.OnboardingRequired,
		CustomerInfoRequired: cs.CustomerInfoRequired,
		NTBEligible:          cs.NTBEligible,
		AvailableLimit:       cs.AvailableLimit,
		Status:               deriveStatus(cs, elig),
	}
	if elig != nil {
		txnElig := elig.TxnEligible
		resp.TxnEligible = &txnElig
		resp.EmiPlans = elig.EmiPlans
		existingUser := elig.ExistingUser
		resp.ExistingUser = &existingUser
		resp.ReasonCode = elig.ReasonCode
		resp.ReasonMessage = elig.ReasonMessage
		if elig.AvailableLimit > 0 {
			resp.AvailableLimit = elig.AvailableLimit
		}
	}
	return resp
}

func deriveStatus(cs *profileResp.CustomerStatusResponse, elig *profileResp.EligibilityResponse) string {
	if cs.OnboardingRequired {
		if cs.NTBEligible != nil && *cs.NTBEligible {
			return "NTB"
		}
		return "NOT_STARTED"
	}
	if cs.PreApproved && cs.AvailableLimit > 0 {
		return "ACTIVE"
	}
	return "INELIGIBLE"
}
