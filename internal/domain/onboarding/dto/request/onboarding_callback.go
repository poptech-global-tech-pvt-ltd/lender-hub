package request

type OnboardingCallbackRequest struct {
	UserID       string  `json:"userId" binding:"required"`
	MerchantID   string  `json:"merchantId" binding:"required"`
	Provider     string  `json:"provider" binding:"required"`
	OnboardingID string  `json:"onboardingId" binding:"required"`
	Mobile       string  `json:"mobile" binding:"required"`
	EventType    string  `json:"eventType" binding:"required"`
	Status       string  `json:"status" binding:"required"`
	Step         *string `json:"step"`
	ErrorCode    *string `json:"errorCode"`
	Message      *string `json:"message"`
	RawEventID   string  `json:"rawEventId"`
	EventTime    string  `json:"eventTime" binding:"required"`
}
