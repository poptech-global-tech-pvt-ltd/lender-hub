package request

// EligibilityRequest is for POST /v1/payin3/eligibility
type EligibilityRequest struct {
	UserID   string  `json:"userId" binding:"required"`
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Currency string  `json:"currency" binding:"required"`
	Source   string  `json:"source" binding:"required"`
}
