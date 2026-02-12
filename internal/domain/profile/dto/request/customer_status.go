package request

type CustomerStatusRequest struct {
	UserID     string         `json:"userId" binding:"required"`
	Mobile     string         `json:"mobile" binding:"required"`
	Email      string         `json:"email"`
	MerchantID string         `json:"merchantId" binding:"required"`
	Source     string         `json:"source" binding:"required"`
	Context    RequestContext `json:"context"`
}

type RequestContext struct {
	Platform string `json:"platform"`
	DeviceID string `json:"deviceId"`
	IP       string `json:"ip"`
}
