package request

import "lending-hub-service/internal/adapter/lazypay/dto/common"

// LPOnboardingRequest matches Lazypay /v7/createStandaloneOnboarding
type LPOnboardingRequest struct {
	AccessKey      string              `json:"accessKey"`
	MerchantID     string              `json:"merchantId"`
	MerchantTxnID  string              `json:"merchantTxnId"`
	User           common.LPUserDetails `json:"user"`
	Address        *common.LPAddress   `json:"address,omitempty"`
	ReturnURL      string              `json:"returnUrl"`
	Channel        string              `json:"channel"`
	PAN            string              `json:"pan,omitempty"`
	Gender         string              `json:"gender,omitempty"`
	DOB            string              `json:"dob,omitempty"`
	FullLegalName  string              `json:"fullLegalName,omitempty"`
	EmploymentType string              `json:"employmentType,omitempty"`
	MonthlySalary  float64             `json:"monthlySalary,omitempty"`
	CompanyName    string              `json:"companyName,omitempty"`
	CompanyType    string              `json:"companyType,omitempty"`
	BureauConsent  bool                `json:"bureauPullConsent"`
	FatherName     string              `json:"fatherName,omitempty"`
	MaritalStatus  bool                `json:"maritalStatus,omitempty"`
	Education      string              `json:"educationalQualification,omitempty"`
	SubMerchantID  string              `json:"subMerchantId,omitempty"`
}
