package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	refundReq "lending-hub-service/internal/domain/refund/dto/request"
	refundResp "lending-hub-service/internal/domain/refund/dto/response"
	"lending-hub-service/internal/adapter/lazypay/config"
	lpConstants "lending-hub-service/internal/adapter/lazypay/constants"
	lpResp "lending-hub-service/internal/adapter/lazypay/dto/response"
	"lending-hub-service/internal/adapter/lazypay/mapper"
	"lending-hub-service/internal/adapter/lazypay/signature"
	"lending-hub-service/internal/infrastructure/http/executor"
	sharedErrors "lending-hub-service/internal/shared/errors"
)

// RefundClient implements RefundGateway for Lazypay
type RefundClient struct {
	config   *config.LazypayConfig
	signer   *signature.SignatureService
	executor executor.HttpExecutor
}

// NewRefundClient creates a new RefundClient
func NewRefundClient(
	cfg *config.LazypayConfig,
	signer *signature.SignatureService,
	exec executor.HttpExecutor,
) *RefundClient {
	return &RefundClient{
		config:   cfg,
		signer:   signer,
		executor: exec,
	}
}

// ProcessRefund implements RefundGateway.ProcessRefund
func (c *RefundClient) ProcessRefund(ctx context.Context, req refundReq.CreateRefundRequest) (*refundResp.RefundResponse, error) {
	// Sign request (using paymentId as merchantTxnId)
	sig := c.signer.SignOrder(req.PaymentID, req.Amount) // Reuse order signature format

	// Map to LP request
	lpReq := mapper.ToLPRefundRequest(req, req.PaymentID, c.config.AccessKey, c.config.MerchantID, sig)

	// Marshal to JSON
	jsonBody, err := json.Marshal(lpReq)
	if err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to marshal request: "+err.Error())
	}

	// Build executor request
	execReq := executor.Request{
		Method: http.MethodPost,
		URL:    c.config.BaseURL + lpConstants.PathRefund,
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
	var lpResp lpResp.LPRefundResponse
	if err := json.Unmarshal(resp.Body, &lpResp); err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to unmarshal response: "+err.Error())
	}

	// Map to canonical response
	return mapper.FromLPRefundResponse(&lpResp, req.RefundID, req.PaymentID, req.Amount, req.Currency), nil
}

// handleErrorResponse parses error response and returns DomainError
func (c *RefundClient) handleErrorResponse(body []byte) (*refundResp.RefundResponse, error) {
	var lpError struct {
		ErrorCode    string `json:"errorCode"`
		ErrorMessage string `json:"errorMessage"`
	}
	if err := json.Unmarshal(body, &lpError); err == nil && lpError.ErrorCode != "" {
		return nil, mapper.MapLPError(lpError.ErrorCode)
	}
	return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "provider error")
}
