package port

import "context"

// ProfileUpdater defines the profile update operations required by the order module.
// Implemented by profile.Service.ProfileUpdater.
type ProfileUpdater interface {
	UpdateLimit(ctx context.Context, userID, lender string, newAvailable float64) error
}
