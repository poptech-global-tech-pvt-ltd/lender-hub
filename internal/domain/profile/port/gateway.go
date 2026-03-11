package port

import "context"

// ProfileGateway abstracts external lender eligibility and customer status calls
type ProfileGateway interface {
	CheckEligibility(ctx context.Context, mobile, email string, amount float64) (*EligibilityResult, error)
	GetCustomerStatus(ctx context.Context, mobile, email string) (*CustomerStatusResult, error)
}

// EligibilityResult is the gateway result from CheckEligibility (Lazypay eligibility API)
type EligibilityResult struct {
	TxnEligible           bool
	EligibilityCode       string
	EligibilityReason     string
	AvailableLimit        float64
	EmiPlans              []EmiPlanResult
	EligibilityResponseID string
	ExistingUser          bool
}

// CustomerStatusResult is the gateway result from GetCustomerStatus (Lazypay customer-status API)
type CustomerStatusResult struct {
	PreApproved        bool
	OnboardingRequired bool
	NTBEligible        *bool
	AvailableLimit     float64
}

// EmiPlanResult represents an EMI plan from gateway
type EmiPlanResult struct {
	Tenure             int
	EMI                float64
	InterestRate       float64
	Principal          float64
	TotalPayableAmount float64
	FirstEmiDueDate    string
	Type               string
}
