package response

// LPRefundResponse matches Lazypay refund success response
type LPRefundResponse struct {
	Status      string `json:"status"`      // REFUND_SUCCESS
	LpTxnID     string `json:"lpTxnId"`     // REFUND transaction lpTxnId
	ParentTxnID string `json:"parentTxnId"` // SALE transaction lpTxnId
	RespMessage string `json:"respMessage"`
	RefundID    string `json:"refundId,omitempty"`
	LenderRefID string `json:"lenderRefId,omitempty"`
	Message     string `json:"message,omitempty"`
}

// LPRefundErrorResponse matches Lazypay 400 error envelope
type LPRefundErrorResponse struct {
	Timestamp int64  `json:"timestamp"`
	Status    int    `json:"status"`
	Error     string `json:"error"`
	Message   string `json:"message"`
	Path      string `json:"path"`
	ErrorCode string `json:"errorCode"`
}
