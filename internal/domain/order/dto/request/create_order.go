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

// CreateOrderRequest - paymentId is POP's ID (required), stored as payment_id for primary polling
type CreateOrderRequest struct {
	PaymentID  string  `json:"paymentId"  binding:"required"` // POP's ID e.g. "POP_PAY_001"
	UserID     string  `json:"userId"     binding:"required"`
	MerchantID string  `json:"merchantId"`                    // optional; defaults to config
	Amount     float64 `json:"amount"     binding:"required,gt=0"`
	Currency   string  `json:"currency"   binding:"required"`
	Source     string  `json:"source"`                        // CHECKOUT|PDP|PLP|CART|CX; defaults to CHECKOUT
	ReturnURL  string  `json:"returnUrl"  binding:"required"`
	EMIPlan    EMIPlanInput `json:"emiPlan"`                  // optional if EmiSelection provided (legacy)
	EmiSelection *EmiSelection `json:"emiSelection,omitempty"`
}
