package request

import "lending-hub-service/internal/adapter/lazypay/dto/common"

// LPCreateOrderRequest matches Lazypay POST create order (merchantTxnId, amount, userDetails, source, returnUrl)
type LPCreateOrderRequest struct {
	MerchantTxnID string               `json:"merchantTxnId"`
	Amount        common.LPAmount      `json:"amount"`
	UserDetails   common.LPUserDetails `json:"userDetails"`
	Source        string               `json:"source"`
	ReturnURL     string               `json:"returnUrl"`
}
