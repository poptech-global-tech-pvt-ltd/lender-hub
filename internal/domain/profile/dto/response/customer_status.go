package response

type PayIn3Status string

const (
	StatusNotStarted PayIn3Status = "NOT_STARTED"
	StatusInProgress PayIn3Status = "IN_PROGRESS"
	StatusActive     PayIn3Status = "ACTIVE"
	StatusIneligible PayIn3Status = "INELIGIBLE"
	StatusBlocked    PayIn3Status = "BLOCKED"
)

type EmiPlan struct {
	TenureMonths int     `json:"tenureMonths"`
	EmiAmount    float64 `json:"emiAmount"`
	TotalAmount  float64 `json:"totalAmount"`
}

type CustomerStatusResponse struct {
	UserID             string       `json:"userId"`
	Provider           string       `json:"provider"`
	PreApproved        bool         `json:"preApproved"`
	AvailableLimit     float64      `json:"availableLimit"`
	CreditLineActive   bool         `json:"creditLineActive"`
	OnboardingRequired bool         `json:"onboardingRequired"`
	Status             PayIn3Status `json:"status"`
	ReasonCode         string       `json:"reasonCode"`
	ReasonMessage      string       `json:"reasonMessage"`
	EmiPlans           []EmiPlan    `json:"emiPlans,omitempty"`
}
