package entity

type ProfileStatus string

const (
	ProfileNotStarted  ProfileStatus = "NOT_STARTED"
	ProfileInProgress  ProfileStatus = "IN_PROGRESS"
	ProfileActive      ProfileStatus = "ACTIVE"
	ProfileIneligible  ProfileStatus = "INELIGIBLE"
	ProfileBlocked     ProfileStatus = "BLOCKED"
)

// CanTransitionTo validates allowed state transitions
// NOT_STARTED → IN_PROGRESS, ACTIVE, INELIGIBLE
// IN_PROGRESS → ACTIVE, INELIGIBLE, BLOCKED
// ACTIVE      → BLOCKED, INELIGIBLE
// BLOCKED     → ACTIVE, INELIGIBLE
// INELIGIBLE  → IN_PROGRESS (retry)
func (s ProfileStatus) CanTransitionTo(next ProfileStatus) bool {
	transitions := map[ProfileStatus][]ProfileStatus{
		ProfileNotStarted: {ProfileInProgress, ProfileActive, ProfileIneligible},
		ProfileInProgress: {ProfileActive, ProfileIneligible, ProfileBlocked},
		ProfileActive:     {ProfileBlocked, ProfileIneligible},
		ProfileBlocked:    {ProfileActive, ProfileIneligible},
		ProfileIneligible: {ProfileInProgress},
	}
	for _, allowed := range transitions[s] {
		if allowed == next {
			return true
		}
	}
	return false
}
