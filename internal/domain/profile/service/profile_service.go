package service

import (
	"context"

	req "lending-hub-service/internal/domain/profile/dto/request"
	res "lending-hub-service/internal/domain/profile/dto/response"
	"lending-hub-service/internal/domain/profile/port"
)

// ProfileService handles customer status checks
type ProfileService interface {
	GetCustomerStatus(ctx context.Context, req req.CustomerStatusRequest) (*res.CustomerStatusResponse, error)
}

type profileService struct {
	repo  port.ProfileRepository
	gateway port.ProfileGateway
	cache  port.ProfileCache
}

// NewProfileService creates a new ProfileService
func NewProfileService(repo port.ProfileRepository, gw port.ProfileGateway, cache port.ProfileCache) ProfileService {
	return &profileService{
		repo:    repo,
		gateway: gw,
		cache:   cache,
	}
}

// GetCustomerStatus retrieves customer status with cache-first strategy
func (s *profileService) GetCustomerStatus(ctx context.Context, req req.CustomerStatusRequest) (*res.CustomerStatusResponse, error) {
	// Try cache first
	lender := "LAZYPAY" // TODO: make configurable
	cached, found, err := s.cache.Get(ctx, req.UserID, lender)
	if err != nil {
		// Log error but continue to gateway
	}
	if found && cached != nil {
		return cached, nil
	}

	// Cache miss - call gateway
	response, err := s.gateway.CheckEligibility(ctx, req)
	if err != nil {
		return nil, err
	}

	// Cache the response
	if cacheErr := s.cache.Set(ctx, req.UserID, lender, response); cacheErr != nil {
		// Log error but don't fail the request
	}

	return response, nil
}
