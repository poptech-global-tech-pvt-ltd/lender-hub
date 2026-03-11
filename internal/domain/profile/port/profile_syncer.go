package port

import (
	"context"

	"lending-hub-service/internal/infrastructure/userprofile"
)

// ProfileSyncer syncs profile updates to User Profile Service (narrow interface)
type ProfileSyncer interface {
	UpdateLenderProfile(ctx context.Context, req userprofile.LenderProfileUpdateRequest) error
}
