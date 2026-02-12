package port

import (
	"context"
	res "lending-hub-service/internal/domain/profile/dto/response"
)

// ProfileCache abstracts short-TTL caching for eligibility responses
type ProfileCache interface {
	Get(ctx context.Context, userID, lender string) (*res.CustomerStatusResponse, bool, error)
	Set(ctx context.Context, userID, lender string, value *res.CustomerStatusResponse) error
	Invalidate(ctx context.Context, userID, lender string) error
}
