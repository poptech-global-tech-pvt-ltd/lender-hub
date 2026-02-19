package response

// LPEligibilityResponse is for POST /api/lazypay/v7/payment/eligibility
type LPEligibilityResponse struct {
	BNPL                  *LPBNPLResult  `json:"bnpl"`
	COF                   *LPCOFResult   `json:"cof"`
	EligibilityResponseID string         `json:"eligibilityResponseId"`
	CustomParams          interface{}    `json:"customParams"`
	ExistingUser          bool           `json:"existingUser"`
}

// LPBNPLResult represents BNPL section (ignored for Pay-in-3)
type LPBNPLResult struct {
	TxnEligibility  bool     `json:"txnEligibility"`
	Reason          string   `json:"reason"`
	Code            string   `json:"code"`
	UserEligibility bool     `json:"userEligibility"`
	SignUpModes     []string `json:"signUpModes"`
}

// LPCOFResult represents COF section (used for Pay-in-3)
type LPCOFResult struct {
	TxnEligibility  bool        `json:"txnEligibility"`
	Reason          string      `json:"reason"`
	Code            string      `json:"code"`
	AvailableLimit  float64     `json:"availableLimit"`
	EmiPlans        []LPEmiPlan `json:"emiPlans"`
}

// LPEmiPlan represents an EMI plan from eligibility
type LPEmiPlan struct {
	InterestRate             float64  `json:"interestRate"`
	Tenure                   int      `json:"tenure"`
	Emi                      float64  `json:"emi"`
	TotalInterestAmount      float64  `json:"totalInterestAmount"`
	Principal                float64  `json:"principal"`
	TotalProcessingFee       float64  `json:"totalProcessingFee"`
	ProcessingFeeGst         float64  `json:"processingFeeGst"`
	TotalPayableAmount       float64  `json:"totalPayableAmount"`
	FirstEmiDueDate          string   `json:"firstEmiDueDate"`
	SubventionTag            *string  `json:"subventionTag"`
	DiscountedInterestAmount float64  `json:"discountedInterestAmount"`
	Schedule                 *string  `json:"schedule"`
	Type                     string   `json:"type"`
	DownPayment              *float64 `json:"downPayment"`
}
