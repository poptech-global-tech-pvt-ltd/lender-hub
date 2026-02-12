package request

type CreateRefundRequest struct {
	RefundID  string  `json:"refundId" binding:"required"`
	PaymentID string  `json:"paymentId" binding:"required"`
	OrderID   string  `json:"orderId"`
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Currency  string  `json:"currency" binding:"required"`
	Reason    string  `json:"reason" binding:"required"`
}
