package request

// StartOnboardingRequest - onboardingTxnId generated internally, merchantId from config, userContact resolved via UserContactResolver
type StartOnboardingRequest struct {
	UserID    string `json:"userId" binding:"required"`
	Source    string `json:"source" binding:"required"`
	ReturnURL string `json:"returnUrl"` // optional override; falls back to config
}
