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
	refundResp "lending-hub-service/internal/domain/refund/dto/response"
	refundPort "lending-hub-service/internal/domain/refund/port"
	"lending-hub-service/internal/infrastructure/http/executor"
	sharedContext "lending-hub-service/internal/shared/context"
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/pkg/idgen"
	baseLogger "lending-hub-service/pkg/logger"
)

// RefundClient implements RefundGateway for Lazypay
type RefundClient struct {
	config   *config.LazypayConfig
	signer   *signature.SignatureService
	executor executor.HttpExecutor
	logger   *baseLogger.Logger
	mapper   *mapper.RefundMapper
}

// NewRefundClient creates a new RefundClient
func NewRefundClient(
	cfg *config.LazypayConfig,
	signer *signature.SignatureService,
	exec executor.HttpExecutor,
	logger *baseLogger.Logger,
	idgen *idgen.Generator,
) *RefundClient {
	return &RefundClient{
		config:   cfg,
		signer:   signer,
		executor: exec,
		logger:   logger,
		mapper:   mapper.NewRefundMapper(idgen),
	}
}

// ProcessRefund implements RefundGateway.ProcessRefund
func (c *RefundClient) ProcessRefund(ctx context.Context, merchantTxnID string, amount float64, currency string) (*refundResp.RefundResponse, string, error) {
	// Extract RequestContext
	rc := sharedContext.FromContext(ctx)

	// Map to LP request and generate refundTxnId (Postman contract: merchantTxnId, amount, refundTxnId)
	lpReq, refundTxnID := c.mapper.ToLPRequest(merchantTxnID, amount, currency)

	// Marshal to JSON
	jsonBody, err := json.Marshal(lpReq)
	if err != nil {
		return nil, "", sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to marshal request: "+err.Error())
	}

	// Sign request (NO email in signature for refunds)
	sig := c.signer.SignRefund(merchantTxnID, amount)

	// Build executor request
	url := fmt.Sprintf("%s%s", c.config.BaseURL, lpConstants.PathRefund)
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
		return nil, "", fmt.Errorf("executor error: %w", err)
	}

	// Log response
	logLazypayResponse(c.logger, ctx, execReq.URL, resp.StatusCode, resp.Body, nil)

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		r, err := c.handleErrorResponse(resp.Body)
		return r, "", err
	}

	// Unmarshal response
	var lpResp lpResp.LPRefundResponse
	if err := json.Unmarshal(resp.Body, &lpResp); err != nil {
		return nil, "", sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to unmarshal response: "+err.Error())
	}

	// Map to canonical response (RefundID/PaymentID filled by service layer; gateway has no caller context)
	return mapper.FromLPRefundResponse(&lpResp, "", "", amount, currency), refundTxnID, nil
}

// EnquireRefund implements RefundGateway.EnquireRefund
func (c *RefundClient) EnquireRefund(ctx context.Context, merchantTxnID string) (*refundPort.EnquiryResponse, error) {
	// Extract RequestContext
	rc := sharedContext.FromContext(ctx)

	// Generate signature for enquiry
	sig := c.signer.SignEnquiry(merchantTxnID)

	// Build URL with query param (Postman: merchantTxnId)
	url := fmt.Sprintf("%s%s?merchantTxnId=%s", c.config.BaseURL, lpConstants.PathRefundEnquiry, merchantTxnID)

	// Build executor request
	execReq := executor.Request{
		Method:  http.MethodGet,
		URL:     url,
		Headers: headersWithoutDevice(c.config.AccessKey, sig, rc), // NO deviceId
		Body:    nil,
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
		return nil, fmt.Errorf("enquiry API failed with status %d", resp.StatusCode)
	}

	// Unmarshal response
	var lpResp lpResp.LPEnquiryResponse
	if err := json.Unmarshal(resp.Body, &lpResp); err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to unmarshal enquiry response: "+err.Error())
	}

	// Map to canonical EnquiryResponse
	return mapEnquiryResponse(&lpResp), nil
}

// mapEnquiryResponse converts LPEnquiryResponse to canonical EnquiryResponse
func mapEnquiryResponse(lpResp *lpResp.LPEnquiryResponse) *refundPort.EnquiryResponse {
	result := &refundPort.EnquiryResponse{
		Order: refundPort.EnquiryOrder{
			OrderID: lpResp.Order.OrderID,
			Status:  lpResp.Order.Status,
			Message: lpResp.Order.Message,
		},
		Transactions: make([]refundPort.EnquiryTransaction, len(lpResp.Transactions)),
	}

	for i, txn := range lpResp.Transactions {
		result.Transactions[i] = refundPort.EnquiryTransaction{
			Status:      txn.Status,
			RespMessage: txn.RespMessage,
			LpTxnID:     txn.LpTxnID,
			TxnType:     txn.TxnType,
			TxnRefNo:    txn.TxnRefNo,
			TxnDateTime: txn.TxnDateTime,
			Amount:      txn.Amount,
		}
	}

	return result
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
