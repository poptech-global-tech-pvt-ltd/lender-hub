package response

type StepDetail struct {
	Step        string  `json:"step"`
	Status      string  `json:"status"`
	CompletedAt *string `json:"completedAt,omitempty"`
}

type OnboardingStatusResponse struct {
	OnboardingID           string       `json:"onboardingId"`
	UserID                 string       `json:"userId"`
	Provider               string       `json:"provider"`
	Status                 string       `json:"status"`
	COFEligible            bool         `json:"cofEligible"`
	LastStep               *string      `json:"lastStep"`
	Steps                  []StepDetail `json:"steps"`
	RejectionReasonCode    *string      `json:"rejectionReasonCode"`
	RejectionReasonMessage *string      `json:"rejectionReasonMessage"`
	Retrying               bool         `json:"retrying"`
	RetryCount             int          `json:"retryCount"`
	NextRetryAt            *string      `json:"nextRetryAt"`
	UpdatedAt              string       `json:"updatedAt"`
}
