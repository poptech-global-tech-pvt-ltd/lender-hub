package request

// CreateRefundRequest - refundId is server-generated, paymentRefundId is POP's refund reference
type CreateRefundRequest struct {
	PaymentID       string  `json:"paymentId" binding:"required"`       // POP's order paymentId
	PaymentRefundID string  `json:"paymentRefundId" binding:"required"` // POP's refund reference
	Amount          float64 `json:"amount" binding:"required,gt=0"`
	Currency        string  `json:"currency" binding:"required"`
	Reason          string  `json:"reason" binding:"required,oneof=USER_CANCELLED PRODUCT_RETURN ORDER_CANCELLED"`
}
