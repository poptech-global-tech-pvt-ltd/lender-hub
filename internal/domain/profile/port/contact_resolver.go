package port

import (
	"context"

	"lending-hub-service/internal/domain/profile/entity"
)

// ContactResolver resolves userId to mobile and email (narrow interface)
type ContactResolver interface {
	GetContact(ctx context.Context, userID string) (*entity.UserContact, error)
}
