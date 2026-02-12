package request

type StartOnboardingRequest struct {
	OnboardingTxnID string       `json:"onboardingTxnId" binding:"required"`
	UserID          string       `json:"userId" binding:"required"`
	MerchantID      string       `json:"merchantId" binding:"required"`
	ChannelID       string       `json:"channelId" binding:"required"`
	Source          string       `json:"source" binding:"required"`
	ReturnURL       string       `json:"returnUrl" binding:"required"`
	UserContact     UserContact  `json:"userContact" binding:"required"`
	Address         Address       `json:"address"`
	KYCSnapshot     KYCSnapshot  `json:"kycSnapshot"`
	Custom          CustomFields `json:"custom"`
}

type UserContact struct {
	Mobile string `json:"mobile" binding:"required"`
	Email  string `json:"email"`
}

type Address struct {
	Street1       string `json:"street1"`
	Street2       string `json:"street2"`
	City          string `json:"city"`
	State         string `json:"state"`
	Country       string `json:"country"`
	Zip           string `json:"zip"`
	Landmark      string `json:"landmark"`
	ResidenceType string `json:"residenceType"`
}

type KYCSnapshot struct {
	PAN                      string            `json:"pan"`
	Gender                   string            `json:"gender"`
	DOB                      string            `json:"dob"`
	FullLegalName            string            `json:"fullLegalName"`
	EmploymentDetails        EmploymentDetails `json:"employmentDetails"`
	BureauPullConsent        bool              `json:"bureauPullConsent"`
	FatherName               string            `json:"fatherName"`
	MaritalStatus            bool              `json:"maritalStatus"`
	EducationalQualification string            `json:"educationalQualification"`
}

type EmploymentDetails struct {
	Type          string  `json:"type"`
	MonthlySalary float64 `json:"monthlySalary"`
	CompanyName   string  `json:"companyName"`
	CompanyType   string  `json:"companyType"`
}

type CustomFields struct {
	SubMerchantID string `json:"subMerchantId"`
}
