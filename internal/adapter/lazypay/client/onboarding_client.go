package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"lending-hub-service/internal/adapter/lazypay/config"
	lpConstants "lending-hub-service/internal/adapter/lazypay/constants"
	lpResp "lending-hub-service/internal/adapter/lazypay/dto/response"
	"lending-hub-service/internal/adapter/lazypay/mapper"
	"lending-hub-service/internal/adapter/lazypay/signature"
	onbResp "lending-hub-service/internal/domain/onboarding/dto/response"
	"lending-hub-service/internal/infrastructure/http/executor"
	sharedContext "lending-hub-service/internal/shared/context"
	sharedErrors "lending-hub-service/internal/shared/errors"
	baseLogger "lending-hub-service/pkg/logger"
)

// OnboardingClient implements OnboardingGateway for Lazypay
type OnboardingClient struct {
	config   *config.LazypayConfig
	signer   *signature.SignatureService
	executor executor.HttpExecutor
	logger   *baseLogger.Logger
	mapper   *mapper.OnboardingMapper
}

// NewOnboardingClient creates a new OnboardingClient
func NewOnboardingClient(
	cfg *config.LazypayConfig,
	signer *signature.SignatureService,
	exec executor.HttpExecutor,
	logger *baseLogger.Logger,
) *OnboardingClient {
	mapperCfg := &mapper.OnboardingMapperConfig{
		SubMerchantID: cfg.SubMerchantID,
		ReturnURL:     cfg.ReturnURL,
	}
	return &OnboardingClient{
		config:   cfg,
		signer:   signer,
		executor: exec,
		logger:   logger,
		mapper:   mapper.NewOnboardingMapper(mapperCfg),
	}
}

// StartOnboarding implements OnboardingGateway.StartOnboarding
func (c *OnboardingClient) StartOnboarding(ctx context.Context, mobile, email string) (*onbResp.OnboardingResponse, error) {
	// Extract RequestContext
	rc := sharedContext.FromContext(ctx)

	// Sign request using SignOnboarding (same as SignCustomerStatus)
	sig := c.signer.SignOnboarding(mobile)

	// Map to LP request (Postman contract: customParams, userDetails, returnUrl, source)
	lpReq := c.mapper.ToLPRequest(mobile, email)

	// Marshal to JSON
	jsonBody, err := json.Marshal(lpReq)
	if err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to marshal request: "+err.Error())
	}

	// Build executor request
	url := fmt.Sprintf("%s%s", c.config.BaseURL, lpConstants.PathOnboarding)
	execReq := executor.Request{
		Method:  http.MethodPost,
		URL:     url,
		Headers: headersWithDevice(c.config.AccessKey, sig, rc),
		Body:    bytes.NewReader(jsonBody),
	}

	// Log request
	logLazypayRequest(c.logger, ctx, execReq.Method, execReq.URL, execReq.Headers, jsonBody)

	// Execute request
	resp, err := c.executor.Do(ctx, execReq)
	if err != nil {
		logLazypayResponse(c.logger, ctx, execReq.URL, 0, nil, fmt.Errorf("executor error: %w", err))
		return nil, fmt.Errorf("executor error: %w", err)
	}

	// Log response
	logLazypayResponse(c.logger, ctx, execReq.URL, resp.StatusCode, resp.Body, nil)

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return nil, c.handleErrorResponse(resp.Body)
	}

	// Unmarshal response
	var lpResp lpResp.LPOnboardingResponse
	if err := json.Unmarshal(resp.Body, &lpResp); err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to unmarshal response: "+err.Error())
	}

	// Map to canonical response (onboardingTxnID generated in service layer, not available here)
	return &onbResp.OnboardingResponse{
		OnboardingID:    lpResp.OnboardingID,
		OnboardingTxnID: "", // Set by service layer
		Provider:        "LAZYPAY",
		RedirectURL:     lpResp.RedirectURL,
		Status:          lpResp.Status,
	}, nil
}

// GetOnboardingStatus implements OnboardingGateway.GetOnboardingStatus
func (c *OnboardingClient) GetOnboardingStatus(ctx context.Context, onboardingID string) (*onbResp.OnboardingStatusResponse, error) {
	// Extract RequestContext
	rc := sharedContext.FromContext(ctx)

	// Build URL with query params
	url := fmt.Sprintf("%s%s?onboardingId=%s", c.config.BaseURL, lpConstants.PathOnboardingStatus, onboardingID)

	// Build executor request
	execReq := executor.Request{
		Method: http.MethodGet,
		URL:    url,
		Headers: map[string]string{
			lpConstants.HeaderAccessKey:     c.config.AccessKey,
			lpConstants.HeaderContentType:   lpConstants.ContentTypeJSON,
			lpConstants.HeaderPlatform:      rc.Platform,
			lpConstants.HeaderUserIPAddress: rc.UserIP,
		},
		Body: nil,
	}

	// Log request
	logLazypayRequest(c.logger, ctx, execReq.Method, execReq.URL, execReq.Headers, nil)

	// Execute request
	resp, err := c.executor.Do(ctx, execReq)
	if err != nil {
		logLazypayResponse(c.logger, ctx, execReq.URL, 0, nil, fmt.Errorf("executor error: %w", err))
		return nil, fmt.Errorf("executor error: %w", err)
	}

	// Log response
	logLazypayResponse(c.logger, ctx, execReq.URL, resp.StatusCode, resp.Body, nil)

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return nil, c.handleErrorResponse(resp.Body)
	}

	// Unmarshal response (assuming same structure as LPOnboardingResponse)
	var lpResp lpResp.LPOnboardingResponse
	if err := json.Unmarshal(resp.Body, &lpResp); err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to unmarshal response: "+err.Error())
	}

	// Map to canonical response
	// Note: OnboardingStatusResponse structure may need adjustment based on actual LP response
	return &onbResp.OnboardingStatusResponse{
		OnboardingID: "", // Not in LP response
		UserID:       "", // Not in LP response
		Provider:     "LAZYPAY",
		Status:       lpResp.Status,
		COFEligible:  lpResp.COFEligible,
		LastStep:     nil, // Not in LP response
		Steps:        nil, // Not in LP response
		UpdatedAt:    "",  // Not in LP response
	}, nil
}

// handleErrorResponse parses error response and returns DomainError
func (c *OnboardingClient) handleErrorResponse(body []byte) error {
	var lpError struct {
		ErrorCode    string `json:"errorCode"`
		ErrorMessage string `json:"errorMessage"`
	}
	if err := json.Unmarshal(body, &lpError); err == nil && lpError.ErrorCode != "" {
		return mapper.MapLPError(lpError.ErrorCode)
	}
	return sharedErrors.New(sharedErrors.CodeInternalError, 500, "provider error")
}
