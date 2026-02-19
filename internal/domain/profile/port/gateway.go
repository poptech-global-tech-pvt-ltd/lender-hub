package port

import (
	"context"
	res "lending-hub-service/internal/domain/profile/dto/response"
)

// ProfileGateway abstracts external lender eligibility calls
type ProfileGateway interface {
	CheckEligibility(ctx context.Context, mobile, email string, amount float64) (*res.EligibilityResponse, error)
	GetCustomerStatus(ctx context.Context, mobile, email string) (*res.CustomerStatusResponse, error)
}
