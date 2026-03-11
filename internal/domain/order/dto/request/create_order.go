package request

// CreateOrderRequest - paymentId is POP's ID (required), stored as payment_id for primary polling.
// MerchantID from config; source defaults to "CHECKOUT".
type CreateOrderRequest struct {
	PaymentID string  `json:"paymentId" binding:"required"` // POP's ID e.g. "POP_PAY_001"
	UserID    string  `json:"userId"    binding:"required"`
	Amount    float64 `json:"amount"    binding:"required,gt=0"`
	Currency  string  `json:"currency"  binding:"required"`
	Source    string  `json:"source"` // defaults to CHECKOUT
	ReturnURL string  `json:"returnUrl" binding:"required"`
}
