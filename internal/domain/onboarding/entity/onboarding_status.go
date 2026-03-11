package entity

type OnboardingStatus string

const (
	OnboardingPending    OnboardingStatus = "PENDING"
	OnboardingInProgress OnboardingStatus = "IN_PROGRESS"
	OnboardingSuccess    OnboardingStatus = "SUCCESS"
	OnboardingIneligible OnboardingStatus = "INELIGIBLE"
	OnboardingFailed     OnboardingStatus = "FAILED"
)

func (s OnboardingStatus) IsTerminal() bool {
	return s == OnboardingSuccess || s == OnboardingIneligible
}
