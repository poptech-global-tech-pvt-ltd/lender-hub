package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"lending-hub-service/internal/adapter/lazypay/config"
	lpConstants "lending-hub-service/internal/adapter/lazypay/constants"
	"lending-hub-service/internal/adapter/lazypay/dto/response"
	"lending-hub-service/internal/adapter/lazypay/mapper"
	"lending-hub-service/internal/adapter/lazypay/signature"
	profileResp "lending-hub-service/internal/domain/profile/dto/response"
	"lending-hub-service/internal/infrastructure/http/executor"
	sharedContext "lending-hub-service/internal/shared/context"
	sharedErrors "lending-hub-service/internal/shared/errors"
	baseLogger "lending-hub-service/pkg/logger"
)

// ProfileClient implements ProfileGateway for Lazypay
type ProfileClient struct {
	config   *config.LazypayConfig
	signer   *signature.SignatureService
	executor executor.HttpExecutor
	logger   *baseLogger.Logger
	mapper   *mapper.ProfileMapper
}

// NewProfileClient creates a new ProfileClient
func NewProfileClient(
	cfg *config.LazypayConfig,
	signer *signature.SignatureService,
	exec executor.HttpExecutor,
	logger *baseLogger.Logger,
) *ProfileClient {
	return &ProfileClient{
		config:   cfg,
		signer:   signer,
		executor: exec,
		logger:   logger,
		mapper:   mapper.NewProfileMapper(),
	}
}

// CheckEligibility implements ProfileGateway.CheckEligibility
func (c *ProfileClient) CheckEligibility(ctx context.Context, mobile, email string, amount float64) (*profileResp.EligibilityResponse, error) {
	// Extract RequestContext
	rc := sharedContext.FromContext(ctx)

	// Map to LP request (Postman contract: userDetails, amount, source)
	lpReq := c.mapper.ToLPEligibilityRequest(mobile, email, amount)

	// Marshal to JSON
	jsonBody, err := json.Marshal(lpReq)
	if err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to marshal request: "+err.Error())
	}

	// Sign request (email conditionally included in signature)
	sig := c.signer.SignEligibility(mobile, email, amount)

	// Build executor request
	url := fmt.Sprintf("%s%s", c.config.BaseURL, lpConstants.PathEligibility)
	execReq := executor.Request{
		Method:  http.MethodPost,
		URL:     url,
		Headers: headersWithDevice(c.config.AccessKey, sig, rc),
		Body:    bytes.NewReader(jsonBody),
	}

	// Log request
	c.logRequest(ctx, execReq.Method, execReq.URL, execReq.Headers, jsonBody)

	// Execute request
	resp, err := c.executor.Do(ctx, execReq)
	if err != nil {
		c.logResponse(ctx, execReq.URL, 0, nil, fmt.Errorf("executor error: %w", err))
		return nil, fmt.Errorf("executor error: %w", err)
	}

	// Log response
	c.logResponse(ctx, execReq.URL, resp.StatusCode, resp.Body, nil)

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return nil, c.handleProfileAPIError(resp.Body)
	}

	// Unmarshal response
	var lpResp response.LPEligibilityResponse
	if err := json.Unmarshal(resp.Body, &lpResp); err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to unmarshal response: "+err.Error())
	}

	// Map to canonical response (userID set by service layer)
	return mapper.FromLPEligibilityResponse(&lpResp, ""), nil
}

// GetCustomerStatus implements ProfileGateway.GetCustomerStatus
func (c *ProfileClient) GetCustomerStatus(ctx context.Context, mobile, email string) (*profileResp.CustomerStatusResponse, error) {
	// Extract RequestContext
	rc := sharedContext.FromContext(ctx)

	// Sign request
	sig := c.signer.SignCustomerStatus(mobile)

	// Build URL with query params (Postman: mobile and merchantId)
	url := fmt.Sprintf("%s%s?mobile=%s&merchantId=%s",
		c.config.BaseURL, lpConstants.PathCustomerStatus, mobile, c.config.SubMerchantID)

	// Build body (Postman: GET with body { userDetails: { mobile, email } })
	body := c.mapper.ToLPCustomerStatusRequest(mobile, email)
	bodyBytes, _ := json.Marshal(body)

	// Build executor request
	execReq := executor.Request{
		Method:  http.MethodGet,
		URL:     url,
		Headers: headersWithoutDevice(c.config.AccessKey, sig, rc), // NO deviceId
		Body:    bytes.NewReader(bodyBytes),                        // GET with body
	}

	// Log request
	c.logRequest(ctx, execReq.Method, execReq.URL, execReq.Headers, bodyBytes)

	// Execute request
	resp, err := c.executor.Do(ctx, execReq)
	if err != nil {
		c.logResponse(ctx, execReq.URL, 0, nil, fmt.Errorf("executor error: %w", err))
		return nil, fmt.Errorf("executor error: %w", err)
	}

	// Log response
	c.logResponse(ctx, execReq.URL, resp.StatusCode, resp.Body, nil)

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return nil, c.handleProfileAPIError(resp.Body)
	}

	// Unmarshal response (Customer Status API has different shape than Eligibility)
	var lpResp response.LPCustomerStatusResponse
	if err := json.Unmarshal(resp.Body, &lpResp); err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to unmarshal response: "+err.Error())
	}

	// Map to canonical response (userID set by service layer)
	return mapper.FromLPCustomerStatusResponse(&lpResp, ""), nil
}

// handleProfileAPIError parses LP error response and returns DomainError
func (c *ProfileClient) handleProfileAPIError(body []byte) error {
	// Try to parse LP error response
	var lpError struct {
		ErrorCode    string `json:"errorCode"`
		ErrorMessage string `json:"errorMessage"`
	}
	if err := json.Unmarshal(body, &lpError); err == nil && lpError.ErrorCode != "" {
		return mapper.MapLPError(lpError.ErrorCode)
	}
	return sharedErrors.New(sharedErrors.CodeInternalError, 500, "provider error")
}

// logRequest logs the outgoing Lazypay request
func (c *ProfileClient) logRequest(ctx context.Context, method, url string, headers map[string]string, body []byte) {
	logLazypayRequest(c.logger, ctx, method, url, headers, body)
}

// logResponse logs the incoming Lazypay response
func (c *ProfileClient) logResponse(ctx context.Context, url string, statusCode int, body []byte, err error) {
	logLazypayResponse(c.logger, ctx, url, statusCode, body, err)
}
