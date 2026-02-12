package response

type OrderStatusResponse struct {
	PaymentID            string  `json:"paymentId"`
	UserID               string  `json:"userId"`
	MerchantID           string  `json:"merchantId"`
	Amount               float64 `json:"amount"`
	Currency             string  `json:"currency"`
	Status               string  `json:"status"`
	LenderOrderID        *string `json:"lenderOrderId,omitempty"`
	LenderMerchantTxnID   *string `json:"lenderMerchantTxnId,omitempty"`
	LenderLastStatus      *string `json:"lenderLastStatus,omitempty"`
	LenderLastTxnID       *string `json:"lenderLastTxnId,omitempty"`
	LenderLastTxnStatus   *string `json:"lenderLastTxnStatus,omitempty"`
	LenderLastTxnMessage  *string `json:"lenderLastTxnMessage,omitempty"`
	LastErrorCode         *string `json:"lastErrorCode,omitempty"`
	LastErrorMessage      *string `json:"lastErrorMessage,omitempty"`
	CreatedAt             string  `json:"createdAt"`
	UpdatedAt             string  `json:"updatedAt"`
}
