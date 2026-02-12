package request

type RefundCallbackRequest struct {
	RefundID      string  `json:"refundId" binding:"required"`
	PaymentID     string  `json:"paymentId" binding:"required"`
	Provider      string  `json:"provider" binding:"required"`
	Status        string  `json:"status" binding:"required"`
	LenderRefID   *string `json:"lenderRefId"`
	LenderMessage *string `json:"lenderMessage"`
	EventTime     string  `json:"eventTime" binding:"required"`
}
