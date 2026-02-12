package request

import "lending-hub-service/internal/adapter/lazypay/dto/common"

// LPCreateOrderRequest matches Lazypay /cof/v0/payment/order
type LPCreateOrderRequest struct {
	AccessKey     string              `json:"accessKey"`
	MerchantID    string              `json:"merchantId"`
	MerchantTxnID string              `json:"merchantTxnId"` // = paymentId
	User          common.LPUserDetails `json:"user"`
	Amount        common.LPAmount     `json:"amount"`
	ReturnURL     string              `json:"returnUrl"`
	Signature     string              `json:"signature"`
	Address       *common.LPAddress   `json:"address,omitempty"`
	EMITenure     int                 `json:"emiTenure"`
	ProductLines  []LPProductLine     `json:"productLines,omitempty"`
}

// LPProductLine represents a product line item
type LPProductLine struct {
	Name     string `json:"name"`
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
	Price    string `json:"price"` // String "1000.00"
}
