package stub

import (
	"context"

	"lending-hub-service/internal/domain/profile/port"
)

// StubProfileGateway implements port.ProfileGateway for local development
type StubProfileGateway struct{}

// NewStubProfileGateway creates a new stub gateway
func NewStubProfileGateway() port.ProfileGateway {
	return &StubProfileGateway{}
}

// CheckEligibility returns a hardcoded EligibilityResult
func (g *StubProfileGateway) CheckEligibility(ctx context.Context, mobile, email string, amount float64) (*port.EligibilityResult, error) {
	return &port.EligibilityResult{
		TxnEligible:       true,
		EligibilityCode:   "COFELIGIBLE",
		EligibilityReason: "",
		AvailableLimit:    50000.0,
		ExistingUser:      true,
		EmiPlans: []port.EmiPlanResult{
			{Tenure: 3, EMI: 16666.67, InterestRate: 0, Principal: 50000, TotalPayableAmount: 50000, FirstEmiDueDate: "2026-03-01", Type: "PAY_IN_PARTS"},
			{Tenure: 6, EMI: 8333.33, InterestRate: 0, Principal: 50000, TotalPayableAmount: 50000, FirstEmiDueDate: "2026-03-01", Type: "PAY_IN_PARTS"},
		},
	}, nil
}

// GetCustomerStatus returns a hardcoded CustomerStatusResult
func (g *StubProfileGateway) GetCustomerStatus(ctx context.Context, mobile, email string) (*port.CustomerStatusResult, error) {
	return &port.CustomerStatusResult{
		PreApproved:        true,
		OnboardingRequired: false,
		AvailableLimit:     50000.0,
		NTBEligible:        nil,
	}, nil
}
