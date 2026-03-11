package response

// LPOrderResponse matches Lazypay order creation response
type LPOrderResponse struct {
	OrderID     string `json:"orderId"`     // Lazypay's order ID
	Status      string `json:"status"`      // "PENDING", "SUCCESS", "FAILED"
	RedirectURL string `json:"redirectUrl"` // User redirect for confirmation
	TxnID       string `json:"txnId,omitempty"`
}
