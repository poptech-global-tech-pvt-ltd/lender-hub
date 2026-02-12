package mapper

import (
	onbReq "lending-hub-service/internal/domain/onboarding/dto/request"
	onbResp "lending-hub-service/internal/domain/onboarding/dto/response"
	lpCommon "lending-hub-service/internal/adapter/lazypay/dto/common"
	lpReq "lending-hub-service/internal/adapter/lazypay/dto/request"
	lpResp "lending-hub-service/internal/adapter/lazypay/dto/response"
)

// ToLPOnboardingRequest converts canonical StartOnboardingRequest → LP format
func ToLPOnboardingRequest(
	req onbReq.StartOnboardingRequest,
	accessKey, merchantID string,
) *lpReq.LPOnboardingRequest {
	lpReq := &lpReq.LPOnboardingRequest{
		AccessKey:     accessKey,
		MerchantID:    merchantID,
		MerchantTxnID: req.OnboardingTxnID,
		User: lpCommon.LPUserDetails{
			Mobile:    req.UserContact.Mobile,
			Email:     req.UserContact.Email,
			FirstName: "", // Not in canonical request
			LastName:  "", // Not in canonical request
		},
		ReturnURL:     req.ReturnURL,
		Channel:       req.ChannelID,
		BureauConsent: req.KYCSnapshot.BureauPullConsent,
	}

	// Map address if provided
	if req.Address.Street1 != "" {
		lpReq.Address = &lpCommon.LPAddress{
			Street1:       req.Address.Street1,
			Street2:       req.Address.Street2,
			City:          req.Address.City,
			State:         req.Address.State,
			Country:       req.Address.Country,
			Zip:           req.Address.Zip,
			Landmark:      req.Address.Landmark,
			ResidenceType: req.Address.ResidenceType,
		}
	}

	// Map KYC fields
	if req.KYCSnapshot.PAN != "" {
		lpReq.PAN = req.KYCSnapshot.PAN
	}
	if req.KYCSnapshot.Gender != "" {
		lpReq.Gender = req.KYCSnapshot.Gender
	}
	if req.KYCSnapshot.DOB != "" {
		lpReq.DOB = req.KYCSnapshot.DOB
	}
	if req.KYCSnapshot.FullLegalName != "" {
		lpReq.FullLegalName = req.KYCSnapshot.FullLegalName
	}
	if req.KYCSnapshot.FatherName != "" {
		lpReq.FatherName = req.KYCSnapshot.FatherName
	}
	if req.KYCSnapshot.EducationalQualification != "" {
		lpReq.Education = req.KYCSnapshot.EducationalQualification
	}

	// Map employment details
	if req.KYCSnapshot.EmploymentDetails.Type != "" {
		lpReq.EmploymentType = req.KYCSnapshot.EmploymentDetails.Type
	}
	if req.KYCSnapshot.EmploymentDetails.MonthlySalary > 0 {
		lpReq.MonthlySalary = req.KYCSnapshot.EmploymentDetails.MonthlySalary
	}
	if req.KYCSnapshot.EmploymentDetails.CompanyName != "" {
		lpReq.CompanyName = req.KYCSnapshot.EmploymentDetails.CompanyName
	}
	if req.KYCSnapshot.EmploymentDetails.CompanyType != "" {
		lpReq.CompanyType = req.KYCSnapshot.EmploymentDetails.CompanyType
	}

	// Map marital status
	if req.KYCSnapshot.MaritalStatus {
		lpReq.MaritalStatus = true
	}

	return lpReq
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
