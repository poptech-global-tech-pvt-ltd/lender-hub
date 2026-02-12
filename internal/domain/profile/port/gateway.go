package port

import (
	"context"
	req "lending-hub-service/internal/domain/profile/dto/request"
	res "lending-hub-service/internal/domain/profile/dto/response"
)

// ProfileGateway abstracts external lender eligibility calls
type ProfileGateway interface {
	CheckEligibility(ctx context.Context, r req.CustomerStatusRequest) (*res.CustomerStatusResponse, error)
	GetCustomerStatus(ctx context.Context, mobile string) (*res.CustomerStatusResponse, error)
}
