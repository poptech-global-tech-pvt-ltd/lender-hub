package response

import "time"

// EligibilityResponse is the canonical response for POST /v1/payin3/eligibility
type EligibilityResponse struct {
	UserID            string    `json:"userId"`
	Lender            string    `json:"lender"`
	TxnEligible       bool      `json:"txnEligible"`
	EligibilityCode   string    `json:"eligibilityCode"`
	EligibilityReason string    `json:"eligibilityReason"`
	AvailableLimit    float64   `json:"availableLimit"`
	CreditLimit       float64   `json:"creditLimit"`
	EmiPlans          []EmiPlan `json:"emiPlans,omitempty"`
	ExistingUser      bool      `json:"existingUser"`
	CheckedAt         time.Time `json:"checkedAt"`
}

// EmiPlan represents an EMI plan option
type EmiPlan struct {
	Tenure             int     `json:"tenure"`
	EMI                float64 `json:"emi"`
	InterestRate       float64 `json:"interestRate"`
	Principal          float64 `json:"principal"`
	TotalPayableAmount float64 `json:"totalPayableAmount"`
	FirstEmiDueDate    string  `json:"firstEmiDueDate"`
	Type               string  `json:"type"`
}
