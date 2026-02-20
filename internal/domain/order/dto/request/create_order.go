package request

// EmiSelection - EMI plan selection from caller (legacy)
type EmiSelection struct {
	Tenure int    `json:"tenure" binding:"required"`
	Type   string `json:"type" binding:"required"` // PAY_IN_PARTS
}

// EMIPlanInput - tenure only for create order (optional if EmiSelection provided)
type EMIPlanInput struct {
	Tenure int `json:"tenure"` // oneof 3,6,9,12; validated in service
}

// CreateOrderRequest - paymentId is server-generated, never accepted from caller
type CreateOrderRequest struct {
	UserID     string  `json:"userId"     binding:"required"`
	MerchantID string  `json:"merchantId"`           // optional; defaults to config
	Amount     float64 `json:"amount"     binding:"required,gt=0"`
	Currency   string  `json:"currency"   binding:"required"`
	Source     string  `json:"source"`               // CHECKOUT|PDP|PLP|CART|CX; defaults to CHECKOUT
	ReturnURL  string  `json:"returnUrl"  binding:"required"`
	EMIPlan    EMIPlanInput `json:"emiPlan"` // optional if EmiSelection provided (legacy)
	// Legacy: caller may send paymentId/emiSelection — we ignore paymentId, use EmiSelection.Tenure if EMIPlan not set
	PaymentID    string        `json:"paymentId,omitempty"`
	EmiSelection *EmiSelection `json:"emiSelection,omitempty"`
}
