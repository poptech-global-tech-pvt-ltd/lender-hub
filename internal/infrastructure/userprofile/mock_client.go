package userprofile

import (
	"context"

	baseLogger "lending-hub-service/pkg/logger"

	"go.uber.org/zap"
)

// MockClient implements ProfileSyncer for upstream sync (TODO: replace with real client when API is ready)
type MockClient struct {
	logger *baseLogger.Logger
}

// NewMockClient creates a new MockClient
func NewMockClient(logger *baseLogger.Logger) *MockClient {
	return &MockClient{logger: logger}
}

// UpdateLenderProfile logs the update (fire-and-forget, non-blocking)
func (c *MockClient) UpdateLenderProfile(ctx context.Context, req LenderProfileUpdateRequest) error {
	c.logger.Info("TODO: upstream profile sync",
		baseLogger.Module("profile"),
		zap.String("userId", req.UserID),
		zap.String("lender", req.Lender),
		zap.Float64("availableLimit", req.AvailableLimit),
		zap.Bool("onboardingDone", req.OnboardingDone),
		zap.String("status", req.CurrentStatus))
	return nil
}
