package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	profileReq "lending-hub-service/internal/domain/profile/dto/request"
	profileResp "lending-hub-service/internal/domain/profile/dto/response"
	"lending-hub-service/internal/adapter/lazypay/config"
	lpConstants "lending-hub-service/internal/adapter/lazypay/constants"
	"lending-hub-service/internal/adapter/lazypay/dto/response"
	"lending-hub-service/internal/adapter/lazypay/mapper"
	"lending-hub-service/internal/adapter/lazypay/signature"
	"lending-hub-service/internal/infrastructure/http/executor"
	sharedErrors "lending-hub-service/internal/shared/errors"
)

// ProfileClient implements ProfileGateway for Lazypay
type ProfileClient struct {
	config   *config.LazypayConfig
	signer   *signature.SignatureService
	executor executor.HttpExecutor
}

// NewProfileClient creates a new ProfileClient
func NewProfileClient(
	cfg *config.LazypayConfig,
	signer *signature.SignatureService,
	exec executor.HttpExecutor,
) *ProfileClient {
	return &ProfileClient{
		config:   cfg,
		signer:   signer,
		executor: exec,
	}
}

// CheckEligibility implements ProfileGateway.CheckEligibility
func (c *ProfileClient) CheckEligibility(ctx context.Context, req profileReq.CustomerStatusRequest) (*profileResp.CustomerStatusResponse, error) {
	// Sign request
	sig := c.signer.SignEligibility(req.Mobile, req.Email, 0) // amount 0 for discovery

	// Map to LP request
	lpReq := mapper.ToLPEligibilityRequest(req, c.config.AccessKey, c.config.MerchantID, sig)

	// Marshal to JSON
	jsonBody, err := json.Marshal(lpReq)
	if err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to marshal request: "+err.Error())
	}

	// Build executor request
	execReq := executor.Request{
		Method: http.MethodPost,
		URL:    c.config.BaseURL + lpConstants.PathEligibility,
		Headers: map[string]string{
			lpConstants.HeaderAccessKey:   c.config.AccessKey,
			lpConstants.HeaderSignature:  sig,
			lpConstants.HeaderContentType: lpConstants.ContentTypeJSON,
		},
		Body: bytes.NewReader(jsonBody),
	}

	// Execute request
	resp, err := c.executor.Do(ctx, execReq)
	if err != nil {
		return nil, fmt.Errorf("executor error: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return c.handleErrorResponse(resp.Body)
	}

	// Unmarshal response
	var lpResp response.LPEligibilityResponse
	if err := json.Unmarshal(resp.Body, &lpResp); err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to unmarshal response: "+err.Error())
	}

	// Map to canonical response
	return mapper.FromLPEligibilityResponse(&lpResp, req.UserID), nil
}

// GetCustomerStatus implements ProfileGateway.GetCustomerStatus
func (c *ProfileClient) GetCustomerStatus(ctx context.Context, mobile string) (*profileResp.CustomerStatusResponse, error) {
	// Sign request
	sig := c.signer.SignCustomerStatus(mobile)

	// Build URL with query params
	url := fmt.Sprintf("%s%s?mobile=%s", c.config.BaseURL, lpConstants.PathCustomerStatus, mobile)

	// Build executor request
	execReq := executor.Request{
		Method: http.MethodGet,
		URL:    url,
		Headers: map[string]string{
			lpConstants.HeaderAccessKey:   c.config.AccessKey,
			lpConstants.HeaderSignature:  sig,
			lpConstants.HeaderContentType: lpConstants.ContentTypeJSON,
		},
		Body: nil,
	}

	// Execute request
	resp, err := c.executor.Do(ctx, execReq)
	if err != nil {
		return nil, fmt.Errorf("executor error: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return c.handleErrorResponse(resp.Body)
	}

	// Unmarshal response
	var lpResp response.LPEligibilityResponse
	if err := json.Unmarshal(resp.Body, &lpResp); err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to unmarshal response: "+err.Error())
	}

	// Map to canonical response (userID not available, use empty string)
	return mapper.FromLPEligibilityResponse(&lpResp, ""), nil
}

// handleErrorResponse parses error response and returns DomainError
func (c *ProfileClient) handleErrorResponse(body []byte) (*profileResp.CustomerStatusResponse, error) {
	// Try to parse LP error response
	var lpError struct {
		ErrorCode    string `json:"errorCode"`
		ErrorMessage string `json:"errorMessage"`
	}
	if err := json.Unmarshal(body, &lpError); err == nil && lpError.ErrorCode != "" {
		return nil, mapper.MapLPError(lpError.ErrorCode)
	}
	// Fallback to generic error
	return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "provider error")
}
