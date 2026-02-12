package stub

import (
	"context"

	req "lending-hub-service/internal/domain/profile/dto/request"
	res "lending-hub-service/internal/domain/profile/dto/response"
	"lending-hub-service/internal/domain/profile/port"
)

// StubProfileGateway implements port.ProfileGateway for local development
type StubProfileGateway struct{}

// NewStubProfileGateway creates a new stub gateway
func NewStubProfileGateway() port.ProfileGateway {
	return &StubProfileGateway{}
}

// CheckEligibility returns a hardcoded ACTIVE response with sample EMI plans
func (g *StubProfileGateway) CheckEligibility(ctx context.Context, r req.CustomerStatusRequest) (*res.CustomerStatusResponse, error) {
	return &res.CustomerStatusResponse{
		UserID:             r.UserID,
		Provider:           "LAZYPAY",
		PreApproved:        true,
		AvailableLimit:     50000.0,
		CreditLineActive:   true,
		OnboardingRequired: false,
		Status:             res.StatusActive,
		ReasonCode:         "",
		ReasonMessage:      "",
		EmiPlans: []res.EmiPlan{
			{
				TenureMonths: 3,
				EmiAmount:    16666.67,
				TotalAmount:  50000.0,
			},
			{
				TenureMonths: 6,
				EmiAmount:    8333.33,
				TotalAmount:  50000.0,
			},
		},
	}, nil
}

// GetCustomerStatus returns the same hardcoded response
func (g *StubProfileGateway) GetCustomerStatus(ctx context.Context, mobile string) (*res.CustomerStatusResponse, error) {
	return &res.CustomerStatusResponse{
		UserID:             "stub-user",
		Provider:           "LAZYPAY",
		PreApproved:        true,
		AvailableLimit:     50000.0,
		CreditLineActive:   true,
		OnboardingRequired: false,
		Status:             res.StatusActive,
		ReasonCode:         "",
		ReasonMessage:      "",
		EmiPlans: []res.EmiPlan{
			{
				TenureMonths: 3,
				EmiAmount:    16666.67,
				TotalAmount:  50000.0,
			},
			{
				TenureMonths: 6,
				EmiAmount:    8333.33,
				TotalAmount:  50000.0,
			},
		},
	}, nil
}
