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
	orderReq "lending-hub-service/internal/domain/order/dto/request"
	orderResp "lending-hub-service/internal/domain/order/dto/response"
	"lending-hub-service/internal/infrastructure/http/executor"
	sharedContext "lending-hub-service/internal/shared/context"
	sharedErrors "lending-hub-service/internal/shared/errors"
	baseLogger "lending-hub-service/pkg/logger"
)

// PaymentClient implements OrderGateway for Lazypay
type PaymentClient struct {
	config   *config.LazypayConfig
	signer   *signature.SignatureService
	executor executor.HttpExecutor
	logger   *baseLogger.Logger
}

// NewPaymentClient creates a new PaymentClient
func NewPaymentClient(
	cfg *config.LazypayConfig,
	signer *signature.SignatureService,
	exec executor.HttpExecutor,
	logger *baseLogger.Logger,
) *PaymentClient {
	return &PaymentClient{
		config:   cfg,
		signer:   signer,
		executor: exec,
		logger:   logger,
	}
}

// CreateOrder implements OrderGateway.CreateOrder
func (c *PaymentClient) CreateOrder(ctx context.Context, req orderReq.CreateOrderRequest) (*orderResp.OrderResponse, error) {
	// Extract RequestContext
	rc := sharedContext.FromContext(ctx)

	// Sign request
	sig := c.signer.SignOrder(req.PaymentID, req.Amount)

	// Map to LP request (use GetMerchantID() which falls back to SubMerchantID)
	lpReq := mapper.ToLPCreateOrderRequest(req, c.config.AccessKey, c.config.GetMerchantID(), sig)

	// Marshal to JSON
	jsonBody, err := json.Marshal(lpReq)
	if err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to marshal request: "+err.Error())
	}

	// Build executor request
	execReq := executor.Request{
		Method: http.MethodPost,
		URL:    c.config.BaseURL + lpConstants.PathCreateOrder,
		Headers: map[string]string{
			lpConstants.HeaderAccessKey:     c.config.AccessKey,
			lpConstants.HeaderSignature:     sig,
			lpConstants.HeaderContentType:   lpConstants.ContentTypeJSON,
			lpConstants.HeaderPlatform:      rc.Platform,
			lpConstants.HeaderUserIPAddress: rc.UserIP,
		},
		Body: bytes.NewReader(jsonBody),
	}

	// Log request
	logLazypayRequest(c.logger, ctx, execReq.Method, execReq.URL, execReq.Headers, jsonBody)

	// Execute request
	resp, err := c.executor.Do(ctx, execReq)
	if err != nil {
		logLazypayResponse(c.logger, ctx, execReq.URL, 0, nil, fmt.Errorf("HTTP executor error: %w", err))
		return nil, fmt.Errorf("HTTP executor error: %w", err)
	}

	// Log response
	logLazypayResponse(c.logger, ctx, execReq.URL, resp.StatusCode, resp.Body, nil)

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return nil, c.handleErrorResponse(resp.Body)
	}

	// Unmarshal response
	var lpResp lpResp.LPOrderResponse
	if err := json.Unmarshal(resp.Body, &lpResp); err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to unmarshal response: "+err.Error())
	}

	// Map to canonical response
	return mapper.FromLPOrderResponse(&lpResp, req.PaymentID), nil
}

// GetOrderStatus implements OrderGateway.GetOrderStatus
func (c *PaymentClient) GetOrderStatus(ctx context.Context, paymentID string) (*orderResp.OrderStatusResponse, error) {
	// Extract RequestContext
	rc := sharedContext.FromContext(ctx)

	// Build URL with query params
	url := fmt.Sprintf("%s%s?merchantTxnId=%s", c.config.BaseURL, lpConstants.PathOrderEnquiry, paymentID)

	// Sign request for enquiry
	sig := c.signer.SignEnquiry(paymentID)

	// Build executor request
	execReq := executor.Request{
		Method: http.MethodGet,
		URL:    url,
		Headers: map[string]string{
			lpConstants.HeaderAccessKey:     c.config.AccessKey,
			lpConstants.HeaderSignature:     sig,
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

	// Unmarshal response
	var lpResp lpResp.LPOrderResponse
	if err := json.Unmarshal(resp.Body, &lpResp); err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to unmarshal response: "+err.Error())
	}

	// Map to canonical response
	lenderOrderID := lpResp.OrderID
	return &orderResp.OrderStatusResponse{
		PaymentID:     paymentID,
		UserID:        "", // Not in LP response
		MerchantID:    "", // Not in LP response
		Amount:        0,  // Not in LP response
		Currency:      "", // Not in LP response
		Status:        lpResp.Status,
		LenderOrderID: &lenderOrderID,
		CreatedAt:     "", // Not in LP response
		UpdatedAt:     "", // Not in LP response
	}, nil
}

// handleErrorResponse parses error response and returns DomainError
func (c *PaymentClient) handleErrorResponse(body []byte) error {
	var lpError struct {
		ErrorCode    string `json:"errorCode"`
		ErrorMessage string `json:"errorMessage"`
	}
	if err := json.Unmarshal(body, &lpError); err == nil && lpError.ErrorCode != "" {
		return mapper.MapLPError(lpError.ErrorCode)
	}
	return sharedErrors.New(sharedErrors.CodeInternalError, 500, "provider error")
}
