package response

type RefundResponse struct {
	RefundID    string  `json:"refundId"`
	PaymentID   string  `json:"paymentId"`
	Provider    string  `json:"provider"`
	Status      string  `json:"status"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	LenderRefID *string `json:"lenderRefId,omitempty"`
}
