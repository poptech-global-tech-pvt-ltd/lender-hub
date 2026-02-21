package port

import "context"

// ProfileUpdater defines the profile update operations required by the refund module.
// Used for restoring credit limit after refund. Implemented by profile.Service.ProfileUpdater.
type ProfileUpdater interface {
	AddToLimit(ctx context.Context, userID, lender string, amount float64) error
}
