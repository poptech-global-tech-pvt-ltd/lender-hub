package webhook

// LPOnboardingWebhook matches Lazypay onboarding webhook payload
type LPOnboardingWebhook struct {
	OnboardingID string  `json:"onboardingId"`
	Mobile       string  `json:"mobile"`
	EventType    string  `json:"eventType"`
	Status       string  `json:"status"`
	Step         *string `json:"step"`
	ErrorCode    *string `json:"errorCode"`
	Message      *string `json:"message"`
	EventTime    string  `json:"eventTime"`
}
