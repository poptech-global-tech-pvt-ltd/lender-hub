package port

import "context"

// ContactInfo holds mobile and email required for order creation.
// Consumer-owned type to avoid depending on profile module entities.
type ContactInfo struct {
	Mobile string
	Email  string
}

// ContactResolver resolves userID to mobile + email.
// Implemented by an adapter wrapping profile.UserContactResolver.
type ContactResolver interface {
	Resolve(ctx context.Context, userID string) (*ContactInfo, error)
	// RefreshFromSource forces a fresh fetch from external source and updates local DB
	RefreshFromSource(ctx context.Context, userID, source string) (*ContactInfo, error)
}
