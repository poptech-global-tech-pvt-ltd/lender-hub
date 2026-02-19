package response

// EligibilityResponse is for POST /v1/payin3/eligibility (txn-level)
type EligibilityResponse struct {
	UserID            string    `json:"userId"`
	Provider          string    `json:"provider"`
	TxnEligible       bool      `json:"txnEligible"`
	AvailableLimit    float64   `json:"availableLimit"`
	EmiPlans          []EmiPlan `json:"emiPlans,omitempty"`
	ExistingUser      bool      `json:"existingUser"`
	ReasonCode        string    `json:"reasonCode,omitempty"`
	ReasonMessage     string    `json:"reasonMessage,omitempty"`
	EligibilityRespID string    `json:"eligibilityResponseId,omitempty"`
}

// EmiPlan represents an EMI plan option
type EmiPlan struct {
	Tenure             int     `json:"tenure"`
	Emi                float64 `json:"emi"`
	InterestRate       float64 `json:"interestRate"`
	Principal          float64 `json:"principal"`
	TotalPayableAmount float64 `json:"totalPayableAmount"`
	FirstEmiDueDate    string  `json:"firstEmiDueDate"`
	Type               string  `json:"type"`
}
