package request

import "lending-hub-service/internal/adapter/lazypay/dto/common"

// LPRefundRequest matches Lazypay /v7/refund
type LPRefundRequest struct {
	AccessKey     string          `json:"accessKey"`
	MerchantID    string          `json:"merchantId,omitempty"` // Optional - not required by Lazypay
	MerchantTxnID string          `json:"merchantTxnId"` // original paymentId
	RefundTxnID   string          `json:"refundTxnId"`   // = refundId
	Amount        common.LPAmount `json:"amount"`
	Reason        string          `json:"reason,omitempty"`
	Signature     string          `json:"signature"`
}
