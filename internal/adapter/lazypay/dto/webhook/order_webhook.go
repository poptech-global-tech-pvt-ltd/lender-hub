package webhook

// LPOrderWebhook matches Lazypay order webhook payload
type LPOrderWebhook struct {
	MerchantTxnID string  `json:"merchantTxnId"`
	OrderID       string  `json:"orderId"`
	Status        string  `json:"status"`
	TxnID         *string `json:"txnId"`
	ErrorCode     *string `json:"errorCode"`
	ErrorMessage  *string `json:"errorMessage"`
	EventTime     string  `json:"eventTime"`
}
