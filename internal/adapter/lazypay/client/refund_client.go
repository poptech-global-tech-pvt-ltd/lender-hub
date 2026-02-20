package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"lending-hub-service/internal/adapter/lazypay/config"
	lpConstants "lending-hub-service/internal/adapter/lazypay/constants"
	lpResp "lending-hub-service/internal/adapter/lazypay/dto/response"
	"lending-hub-service/internal/adapter/lazypay/mapper"
	"lending-hub-service/internal/adapter/lazypay/signature"
	refundPort "lending-hub-service/internal/domain/refund/port"
	"lending-hub-service/internal/infrastructure/http/executor"
	sharedContext "lending-hub-service/internal/shared/context"
	sharedErrors "lending-hub-service/internal/shared/errors"
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
) *RefundClient {
	return &RefundClient{
		config:   cfg,
		signer:   signer,
		executor: exec,
		logger:   logger,
		mapper:   mapper.NewRefundMapper(),
	}
}

// ProcessRefund implements RefundGateway.ProcessRefund — single attempt, no retry
func (c *RefundClient) ProcessRefund(ctx context.Context, req refundPort.ProcessRefundRequest) (*refundPort.ProcessRefundResponse, error) {
	rc := sharedContext.FromContext(ctx)

	lpReq := c.mapper.ToLPRequest(req.MerchantTxnID, req.Amount, req.Currency, req.RefundTxnID)
	jsonBody, err := json.Marshal(lpReq)
	if err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to marshal request: "+err.Error())
	}

	sig := c.signer.SignRefund(req.MerchantTxnID, req.Amount)
	url := fmt.Sprintf("%s%s", c.config.BaseURL, lpConstants.PathRefund)
	execReq := executor.Request{
		Method:  http.MethodPost,
		URL:     url,
		Headers: headersWithDevice(c.config.AccessKey, sig, rc),
		Body:    bytes.NewReader(jsonBody),
	}

	logLazypayRequest(c.logger, ctx, execReq.Method, execReq.URL, execReq.Headers, jsonBody)

	resp, err := c.executor.Do(ctx, execReq)
	if err != nil {
		logLazypayResponse(c.logger, ctx, execReq.URL, 0, nil, fmt.Errorf("executor error: %w", err))
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return &refundPort.ProcessRefundResponse{IsTimeout: true}, nil
		}
		return nil, fmt.Errorf("executor error: %w", err)
	}

	logLazypayResponse(c.logger, ctx, execReq.URL, resp.StatusCode, resp.Body, nil)

	if resp.StatusCode >= 400 {
		return c.handleProcessRefundError(resp.StatusCode, resp.Body)
	}

	var lpResp lpResp.LPRefundResponse
	if err := json.Unmarshal(resp.Body, &lpResp); err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to unmarshal response: "+err.Error())
	}

	respMsg := lpResp.RespMessage
	if respMsg == "" {
		respMsg = lpResp.Message
	}
	return &refundPort.ProcessRefundResponse{
		Status:      lpResp.Status,
		LpTxnID:     lpResp.LpTxnID,
		ParentTxnID: lpResp.ParentTxnID,
		RespMessage: respMsg,
	}, nil
}

func (c *RefundClient) handleProcessRefundError(statusCode int, body []byte) (*refundPort.ProcessRefundResponse, error) {
	var errResp lpResp.LPRefundErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.ErrorCode == "LPDUPLICATEREFUND" {
		return &refundPort.ProcessRefundResponse{ErrorCode: "LPDUPLICATEREFUND"}, nil
	}
	// Try alternate error envelope
	var altErr struct {
		ErrorCode string `json:"errorCode"`
	}
	if err := json.Unmarshal(body, &altErr); err == nil && altErr.ErrorCode == "LPDUPLICATEREFUND" {
		return &refundPort.ProcessRefundResponse{ErrorCode: "LPDUPLICATEREFUND"}, nil
	}
	return nil, sharedErrors.New(sharedErrors.CodeInternalError, statusCode, "provider refund error")
}

// EnquireRefund implements RefundGateway.EnquireRefund
func (c *RefundClient) EnquireRefund(ctx context.Context, merchantTxnID string) (*refundPort.EnquiryResponse, error) {
	rc := sharedContext.FromContext(ctx)
	sig := c.signer.SignEnquiry(merchantTxnID)
	url := fmt.Sprintf("%s%s?merchantTxnId=%s", c.config.BaseURL, lpConstants.PathRefundEnquiry, merchantTxnID)
	execReq := executor.Request{
		Method:  http.MethodGet,
		URL:     url,
		Headers: headersWithoutDevice(c.config.AccessKey, sig, rc),
		Body:    nil,
	}

	logLazypayRequest(c.logger, ctx, execReq.Method, execReq.URL, execReq.Headers, nil)

	resp, err := c.executor.Do(ctx, execReq)
	if err != nil {
		logLazypayResponse(c.logger, ctx, execReq.URL, 0, nil, fmt.Errorf("executor error: %w", err))
		return nil, fmt.Errorf("executor error: %w", err)
	}

	logLazypayResponse(c.logger, ctx, execReq.URL, resp.StatusCode, resp.Body, nil)

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("enquiry API failed with status %d", resp.StatusCode)
	}

	var lpResp lpResp.LPEnquiryResponse
	if err := json.Unmarshal(resp.Body, &lpResp); err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to unmarshal enquiry response: "+err.Error())
	}

	return mapEnquiryResponse(&lpResp), nil
}

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
