package webhook

// LPRefundWebhook matches Lazypay refund webhook payload
type LPRefundWebhook struct {
	RefundTxnID   string  `json:"refundTxnId"`
	MerchantTxnID string  `json:"merchantTxnId"`
	Status         string  `json:"status"`
	LenderRefID    *string `json:"lenderRefId"`
	Message        *string `json:"message"`
	EventTime      string  `json:"eventTime"`
}
