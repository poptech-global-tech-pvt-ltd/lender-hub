package stub

import (
	"context"

	res "lending-hub-service/internal/domain/profile/dto/response"
	"lending-hub-service/internal/domain/profile/port"
)

// StubProfileGateway implements port.ProfileGateway for local development
type StubProfileGateway struct{}

// NewStubProfileGateway creates a new stub gateway
func NewStubProfileGateway() port.ProfileGateway {
	return &StubProfileGateway{}
}

// CheckEligibility returns a hardcoded EligibilityResponse
func (g *StubProfileGateway) CheckEligibility(ctx context.Context, mobile, email string, amount float64) (*res.EligibilityResponse, error) {
	return &res.EligibilityResponse{
		UserID:         "stub-user",
		Provider:       "LAZYPAY",
		TxnEligible:    true,
		AvailableLimit: 50000.0,
		ExistingUser:   true,
		EmiPlans: []res.EmiPlan{
			{Tenure: 3, Emi: 16666.67, InterestRate: 0, Principal: 50000, TotalPayableAmount: 50000, FirstEmiDueDate: "2026-03-01", Type: "PAY_IN_PARTS"},
			{Tenure: 6, Emi: 8333.33, InterestRate: 0, Principal: 50000, TotalPayableAmount: 50000, FirstEmiDueDate: "2026-03-01", Type: "PAY_IN_PARTS"},
		},
	}, nil
}

// GetCustomerStatus returns a hardcoded CustomerStatusResponse
func (g *StubProfileGateway) GetCustomerStatus(ctx context.Context, mobile, email string) (*res.CustomerStatusResponse, error) {
	return &res.CustomerStatusResponse{
		UserID:               "stub-user",
		Provider:             "LAZYPAY",
		PreApproved:          true,
		OnboardingRequired:   false,
		CustomerInfoRequired: false,
		AvailableLimit:       50000.0,
		NTBEligible:          nil,
	}, nil
}
