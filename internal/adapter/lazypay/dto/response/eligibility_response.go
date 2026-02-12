package response

// LPEligibilityResponse is Lazypay's eligibility API response
type LPEligibilityResponse struct {
	Status                string      `json:"status"`       // "APPROVED", "REJECTED"
	EligibilityResponseID string      `json:"eligibilityResponseId"`
	PreApproved           bool         `json:"preApproved"`
	CreditLimit           float64      `json:"creditLimit"`
	AvailableLimit        float64      `json:"availableLimit"`
	CreditLineActive      bool         `json:"creditLineActive"`
	EMIPlans              []LPEMIPlan  `json:"emiPlans"`
	ReasonCode            string       `json:"reasonCode,omitempty"`
	ReasonMessage         string       `json:"reasonMessage,omitempty"`
	Blocked               bool         `json:"blocked"`
	BlockReason           string       `json:"blockReason,omitempty"`
}

// LPEMIPlan represents an EMI plan option
type LPEMIPlan struct {
	Tenure      int     `json:"tenure"`
	EMI         float64 `json:"emi"`
	TotalAmount float64 `json:"totalAmount"`
}
