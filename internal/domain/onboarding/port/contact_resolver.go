package port

import "context"

// ContactInfo holds user contact details for onboarding
type ContactInfo struct {
	Mobile string
	Email  string
}

// ContactResolver is a narrow interface owned by the onboarding module.
// Used to fetch user contact details during onboarding flow.
type ContactResolver interface {
	GetContact(ctx context.Context, userID string) (*ContactInfo, error)
}
