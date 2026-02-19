package stub

import (
	"context"
	"fmt"

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
func (g *StubOnboardingGateway) StartOnboarding(ctx context.Context, mobile, email string) (*res.OnboardingResponse, error) {
	// Generate a fake provider onboarding ID
	providerOnboardingID := fmt.Sprintf("LP-ONB-%s", mobile)

	// Return fake redirect URL
	redirectURL := fmt.Sprintf("https://stub.lazypay.in/onboarding/%s", providerOnboardingID)

	return &res.OnboardingResponse{
		OnboardingID:    providerOnboardingID,
		OnboardingTxnID: "", // Set by service layer
		Provider:        "LAZYPAY",
		RedirectURL:     redirectURL,
		Status:          "PENDING",
	}, nil
}

// GetOnboardingStatus returns a stub status response
func (g *StubOnboardingGateway) GetOnboardingStatus(ctx context.Context, onboardingID string) (*res.OnboardingStatusResponse, error) {
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
