package stub

import (
	"context"
	"fmt"

	req "lending-hub-service/internal/domain/onboarding/dto/request"
	res "lending-hub-service/internal/domain/onboarding/dto/response"
	"lending-hub-service/internal/domain/onboarding/port"
)

// StubOnboardingGateway implements port.OnboardingGateway for local development
type StubOnboardingGateway struct{}

// NewStubOnboardingGateway creates a new stub gateway
func NewStubOnboardingGateway() port.OnboardingGateway {
	return &StubOnboardingGateway{}
}

// StartOnboarding returns a fake redirect URL and PENDING status
func (g *StubOnboardingGateway) StartOnboarding(ctx context.Context, req req.StartOnboardingRequest) (*res.OnboardingResponse, error) {
	// Generate a fake provider onboarding ID
	providerOnboardingID := fmt.Sprintf("LP-%s", req.OnboardingTxnID)

	// Return fake redirect URL
	redirectURL := fmt.Sprintf("https://stub.lazypay.in/onboarding/%s?returnUrl=%s", providerOnboardingID, req.ReturnURL)

	return &res.OnboardingResponse{
		OnboardingID:    providerOnboardingID,
		OnboardingTxnID: req.OnboardingTxnID,
		Provider:        "LAZYPAY",
		RedirectURL:     redirectURL,
		Status:          "PENDING",
	}, nil
}

// GetOnboardingStatus returns a stub status response
func (g *StubOnboardingGateway) GetOnboardingStatus(ctx context.Context, mobile string) (*res.OnboardingStatusResponse, error) {
	return &res.OnboardingStatusResponse{
		OnboardingID: "stub-onboarding-id",
		UserID:       "stub-user",
		Provider:     "LAZYPAY",
		Status:       "PENDING",
		COFEligible:  false,
		RetryCount:   0,
		Steps:        []res.StepDetail{},
	}, nil
}
