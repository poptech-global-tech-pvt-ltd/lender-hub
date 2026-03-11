package entity

import "time"

// UserContact represents cached user contact information
type UserContact struct {
	ID        int64
	UserID    string // global_user_id
	Mobile    string // 10-digit, no prefix
	Email     string // empty string = no email
	RawPhone  string // original phone with prefix
	Source    string // PROFILE_SERVICE, ONBOARDING, ORDER
	CreatedAt time.Time
	UpdatedAt time.Time
}

// HasEmail returns true if email is non-empty
func (uc *UserContact) HasEmail() bool {
	return uc.Email != ""
}
