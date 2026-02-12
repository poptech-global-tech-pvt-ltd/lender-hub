package mapper

import (
	"fmt"

	refundReq "lending-hub-service/internal/domain/refund/dto/request"
	refundResp "lending-hub-service/internal/domain/refund/dto/response"
	lpCommon "lending-hub-service/internal/adapter/lazypay/dto/common"
	lpReq "lending-hub-service/internal/adapter/lazypay/dto/request"
	lpResp "lending-hub-service/internal/adapter/lazypay/dto/response"
)

// ToLPRefundRequest converts canonical CreateRefundRequest → LP format
func ToLPRefundRequest(
	req refundReq.CreateRefundRequest,
	paymentID, accessKey, merchantID, signature string,
) *lpReq.LPRefundRequest {
	return &lpReq.LPRefundRequest{
		AccessKey:     accessKey,
		MerchantID:    merchantID,
		MerchantTxnID: paymentID,
		RefundTxnID:   req.RefundID,
		Amount: lpCommon.LPAmount{
			Value:    fmt.Sprintf("%.2f", req.Amount),
			Currency: req.Currency,
		},
		Reason:    req.Reason,
		Signature: signature,
	}
}

// FromLPRefundResponse converts LP response → canonical RefundResponse
func FromLPRefundResponse(
	lp *lpResp.LPRefundResponse,
	refundID, paymentID string,
	amount float64,
	currency string,
) *refundResp.RefundResponse {
	var lenderRefID *string
	if lp.LenderRefID != "" {
		lenderRefID = &lp.LenderRefID
	}
	return &refundResp.RefundResponse{
		RefundID:   refundID,
		PaymentID:  paymentID,
		Provider:   "LAZYPAY",
		Status:     lp.Status,
		Amount:     amount,
		Currency:   currency,
		LenderRefID: lenderRefID,
	}
}
