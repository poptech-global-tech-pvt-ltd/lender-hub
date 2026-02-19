package request

// CreateRefundRequest - orderId looked up from paymentId internally, refundTxnId generated internally
type CreateRefundRequest struct {
	RefundID  string  `json:"refundId" binding:"required"`  // caller's reference
	PaymentID string  `json:"paymentId" binding:"required"` // which order to refund
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Currency  string  `json:"currency" binding:"required"`
	Reason    string  `json:"reason" binding:"required"` // USER_CANCELLED, PRODUCT_RETURN, etc
}
