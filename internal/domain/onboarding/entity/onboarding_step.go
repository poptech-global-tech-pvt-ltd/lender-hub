package entity

type OnboardingStep string

const (
	StepUserData     OnboardingStep = "USER_DATA"
	StepEMISelection OnboardingStep = "EMI_SELECTION"
	StepKYC          OnboardingStep = "KYC"
	StepKFS          OnboardingStep = "KFS"
	StepMITC         OnboardingStep = "MITC"
	StepAutopay      OnboardingStep = "AUTOPAY"
)

type StepStatus string

const (
	StepPending StepStatus = "PENDING"
	StepSuccess StepStatus = "SUCCESS"
	StepFailed  StepStatus = "FAILED"
)
