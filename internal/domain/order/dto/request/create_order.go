package request

// EmiSelection - EMI plan selection from caller
type EmiSelection struct {
	Tenure int    `json:"tenure" binding:"required"`
	Type   string `json:"type" binding:"required"` // PAY_IN_PARTS
}

// CreateOrderRequest - mobile resolved via UserContactResolver, merchantId from config, merchantTxnId generated internally
type CreateOrderRequest struct {
	PaymentID    string       `json:"paymentId" binding:"required"`    // caller's idempotency key
	UserID       string       `json:"userId" binding:"required"`
	Amount       float64      `json:"amount" binding:"required,gt=0"`
	Currency     string       `json:"currency" binding:"required"`
	EmiSelection EmiSelection `json:"emiSelection" binding:"required"`
	ReturnURL    string       `json:"returnUrl"` // optional override; falls back to config
}
