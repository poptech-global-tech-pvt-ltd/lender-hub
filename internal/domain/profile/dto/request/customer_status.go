package request

// CustomerStatusRequest - mobile/email resolved via UserContactResolver, merchantId from config, context from middleware
type CustomerStatusRequest struct {
	UserID string `json:"userId" binding:"required"`
	Source string `json:"source" binding:"required"`
}
