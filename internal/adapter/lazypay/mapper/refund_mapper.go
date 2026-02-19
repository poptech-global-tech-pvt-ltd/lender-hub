package mapper

import (
	"lending-hub-service/pkg/idgen"

	refundResp "lending-hub-service/internal/domain/refund/dto/response"
	lpCommon "lending-hub-service/internal/adapter/lazypay/dto/common"
	lpReq "lending-hub-service/internal/adapter/lazypay/dto/request"
	lpResp "lending-hub-service/internal/adapter/lazypay/dto/response"
)

// RefundMapper maps between canonical and Lazypay refund formats
type RefundMapper struct {
	idgen *idgen.Generator
}

// NewRefundMapper creates a new refund mapper
func NewRefundMapper(idgen *idgen.Generator) *RefundMapper {
	return &RefundMapper{idgen: idgen}
}

// ToLPRequest builds refund DTO and returns the generated refundTxnId for DB persistence.
// Postman contract: { merchantTxnId, amount, refundTxnId }
func (m *RefundMapper) ToLPRequest(merchantTxnID string, amount float64, currency string) (*lpReq.LPRefundRequest, string) {
	refundID := m.idgen.RefundID() // e.g. REF_01HQZV8X9P...

	return &lpReq.LPRefundRequest{
		MerchantTxnID: merchantTxnID,
		Amount: lpCommon.LPAmount{
			Value:    lpCommon.FormatAmount(amount),
			Currency: currency,
		},
		RefundTxnID: refundID,
	}, refundID
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
