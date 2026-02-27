package request

type OrderCallbackRequest struct {
	PaymentID     string  `json:"paymentId" binding:"required"`
	Provider      string  `json:"provider" binding:"required"`
	Status        string  `json:"status" binding:"required"`
	LenderOrderID string  `json:"lenderOrderId" binding:"required"`
	LenderTxnID   *string `json:"lenderTxnId"`
	ErrorCode     *string `json:"errorCode"`
	ErrorMessage  *string `json:"errorMessage"`
	EventTime     string  `json:"eventTime" binding:"required"`
}
