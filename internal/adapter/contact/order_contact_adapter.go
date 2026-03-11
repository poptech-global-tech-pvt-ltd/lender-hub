package contact

import (
	"context"

	orderPort "lending-hub-service/internal/domain/order/port"
	profileService "lending-hub-service/internal/domain/profile/service"
)

// OrderContactAdapter adapts profile.UserContactResolver to order.ContactResolver.
// Keeps order module free of profile module dependencies.
type OrderContactAdapter struct {
	inner *profileService.UserContactResolver
}

// NewOrderContactAdapter creates an adapter.
func NewOrderContactAdapter(inner *profileService.UserContactResolver) *OrderContactAdapter {
	return &OrderContactAdapter{inner: inner}
}

// Resolve implements orderPort.ContactResolver.
func (a *OrderContactAdapter) Resolve(ctx context.Context, userID string) (*orderPort.ContactInfo, error) {
	uc, err := a.inner.Resolve(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &orderPort.ContactInfo{Mobile: uc.Mobile, Email: uc.Email}, nil
}

// RefreshFromSource implements orderPort.ContactResolver.
func (a *OrderContactAdapter) RefreshFromSource(ctx context.Context, userID, source string) (*orderPort.ContactInfo, error) {
	uc, err := a.inner.RefreshFromSource(ctx, userID, source)
	if err != nil {
		return nil, err
	}
	if uc == nil {
		return nil, nil
	}
	return &orderPort.ContactInfo{Mobile: uc.Mobile, Email: uc.Email}, nil
}
