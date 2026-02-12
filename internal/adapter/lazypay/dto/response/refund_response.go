package response

// LPRefundResponse matches Lazypay refund response
type LPRefundResponse struct {
	RefundID    string `json:"refundId"`
	Status      string `json:"status"`
	LenderRefID string `json:"lenderRefId,omitempty"`
	Message     string `json:"message,omitempty"`
}
