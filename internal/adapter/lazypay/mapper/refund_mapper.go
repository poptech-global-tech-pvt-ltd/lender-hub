package mapper

import (
	refundResp "lending-hub-service/internal/domain/refund/dto/response"
	lpCommon "lending-hub-service/internal/adapter/lazypay/dto/common"
	lpReq "lending-hub-service/internal/adapter/lazypay/dto/request"
	lpResp "lending-hub-service/internal/adapter/lazypay/dto/response"
)

// RefundMapper maps between canonical and Lazypay refund formats
type RefundMapper struct{}

// NewRefundMapper creates a new refund mapper
func NewRefundMapper() *RefundMapper {
	return &RefundMapper{}
}

// ToLPRequest builds refund DTO with given refundTxnID (our generated refundId)
func (m *RefundMapper) ToLPRequest(merchantTxnID string, amount float64, currency string, refundTxnID string) *lpReq.LPRefundRequest {
	return &lpReq.LPRefundRequest{
		MerchantTxnID: merchantTxnID,
		Amount: lpCommon.LPAmount{
			Value:    lpCommon.FormatAmount(amount),
			Currency: currency,
		},
		RefundTxnID: refundTxnID,
	}
}

// FromLPRefundResponse converts LP response → RefundResponse (legacy / tests)
func FromLPRefundResponse(
	lp *lpResp.LPRefundResponse,
	refundID, paymentID, loanID string,
	amount float64,
	currency string,
) *refundResp.RefundResponse {
	return &refundResp.RefundResponse{
		RefundID:            refundID,
		PaymentRefundID:     "",
		ProviderRefundTxnID: lp.LpTxnID,
		PaymentID:           paymentID,
		LoanID:              loanID,
		Status:              lp.Status,
		Amount:              amount,
		Currency:            currency,
	}
}
