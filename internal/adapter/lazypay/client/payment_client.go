package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"


	"lending-hub-service/internal/adapter/lazypay/config"
	lpConstants "lending-hub-service/internal/adapter/lazypay/constants"
	lpCommon "lending-hub-service/internal/adapter/lazypay/dto/common"
	lpReq "lending-hub-service/internal/adapter/lazypay/dto/request"
	lpResp "lending-hub-service/internal/adapter/lazypay/dto/response"
	"lending-hub-service/internal/adapter/lazypay/mapper"
	"lending-hub-service/internal/adapter/lazypay/signature"
	orderPort "lending-hub-service/internal/domain/order/port"
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
func (c *PaymentClient) CreateOrder(ctx context.Context, input orderPort.OrderInput) (*orderResp.OrderResponse, error) {
	// Extract RequestContext
	rc := sharedContext.FromContext(ctx)

	// Map EMI plans from OrderInput to LP format
	lpEmiPlans := make([]lpReq.LPEmiPlan, len(input.EmiPlans))
	for i, plan := range input.EmiPlans {
		lpEmiPlans[i] = lpReq.LPEmiPlan{
			InterestRate:             plan.InterestRate,
			Tenure:                   plan.Tenure,
			Emi:                      plan.Emi,
			TotalInterestAmount:      plan.TotalInterestAmount,
			Principal:                plan.Principal,
			TotalProcessingFee:       plan.TotalProcessingFee,
			ProcessingFeeGst:         plan.ProcessingFeeGst,
			TotalPayableAmount:       plan.TotalPayableAmount,
			FirstEmiDueDate:          plan.FirstEmiDueDate,
			SubventionTag:            plan.SubventionTag,
			DiscountedInterestAmount: plan.DiscountedInterestAmount,
			Schedule:                 plan.Schedule,
			Type:                     plan.Type,
		}
	}

	// Map to LP request (Postman contract: merchantTxnId, amount, userDetails, source, returnUrl, emiPlans)
	lpReq := &lpReq.LPCreateOrderRequest{
		MerchantTxnID: input.MerchantTxnID,
		Amount: lpCommon.LPAmount{
			Value:    lpCommon.FormatAmount(input.Amount),
			Currency: input.Currency,
		},
		UserDetails: lpCommon.NewLPUserDetails(input.Mobile, input.Email),
		Source:      "website",
		ReturnURL:   c.config.ReturnURL, // From config
		EmiPlans:    lpEmiPlans,
	}

	// Marshal to JSON
	jsonBody, err := json.Marshal(lpReq)
	if err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to marshal request: "+err.Error())
	}

	// Sign request (NO email in signature for orders)
	sig := c.signer.SignOrder(input.MerchantTxnID, input.Amount)

	// Build executor request
	url := fmt.Sprintf("%s%s", c.config.BaseURL, lpConstants.PathCreateOrder)
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

	// Map to canonical response (paymentID set by service layer)
	return &orderResp.OrderResponse{
		PaymentID:     "", // Set by service layer
		Status:        lpResp.Status,
		LenderOrderID: lpResp.OrderID,
		RedirectURL:   lpResp.RedirectURL,
		ErrorCode:     nil,
		ErrorMessage:  nil,
	}, nil
}

// GetOrderStatus implements OrderGateway.GetOrderStatus
func (c *PaymentClient) GetOrderStatus(ctx context.Context, merchantTxnID string) (*orderResp.OrderStatusResponse, error) {
	// Extract RequestContext
	rc := sharedContext.FromContext(ctx)

	// Build URL with query params
	url := fmt.Sprintf("%s%s?merchantTxnId=%s", c.config.BaseURL, lpConstants.PathOrderEnquiry, merchantTxnID)

	// Sign request for enquiry
	sig := c.signer.SignEnquiry(merchantTxnID)

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

	// Map to canonical response (used for enquiry; paymentId filled by caller)
	return &orderResp.OrderStatusResponse{
		PaymentID:     "",
		Status:        lpResp.Status,
		LenderOrderID: lpResp.OrderID,
		Amount:        0,
		Currency:      "INR",
		CreatedAt:     time.Time{},
		UpdatedAt:     time.Time{},
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
