package service

import (
	"context"

	"golang.org/x/sync/errgroup"

	req "lending-hub-service/internal/domain/profile/dto/request"
	res "lending-hub-service/internal/domain/profile/dto/response"
	"lending-hub-service/internal/domain/profile/port"
	"lending-hub-service/internal/infrastructure/userprofile"
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/pkg/lender"
	baseLogger "lending-hub-service/pkg/logger"

	"go.uber.org/zap"
)

// ProfileService handles eligibility and customer status checks
type ProfileService interface {
	CheckEligibility(ctx context.Context, req req.EligibilityRequest) (*res.EligibilityResponse, error)
	GetCustomerStatus(ctx context.Context, req req.CustomerStatusRequest) (*res.CustomerStatusResponse, error)
	GetCombinedProfile(ctx context.Context, userID string, amount *float64, currency, source string) (*res.UserProfileResponse, error)
}

type profileService struct {
	gateway         port.ProfileGateway
	contactResolver *UserContactResolver
	repo            port.ProfileRepository
	profileClient   *userprofile.Client
	logger          *baseLogger.Logger
}

// NewProfileService creates a new ProfileService
func NewProfileService(
	gw port.ProfileGateway,
	contactResolver *UserContactResolver,
	repo port.ProfileRepository,
	profileClient *userprofile.Client,
	logger *baseLogger.Logger,
) ProfileService {
	return &profileService{
		gateway:         gw,
		contactResolver: contactResolver,
		repo:            repo,
		profileClient:   profileClient,
		logger:          logger,
	}
}

// CheckEligibility checks txn-level eligibility
func (s *profileService) CheckEligibility(ctx context.Context, req req.EligibilityRequest) (*res.EligibilityResponse, error) {
	contact, err := s.contactResolver.Resolve(ctx, req.UserID)
	if err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeUserContactNotFound, 422, "Unable to resolve user contact details (mobile required)")
	}

	resp, err := s.gateway.CheckEligibility(ctx, contact.Mobile, contact.Email, req.Amount)
	if err != nil {
		return nil, err
	}
	resp.UserID = req.UserID

	l := lender.Lazypay.String()
	if persistErr := s.repo.UpsertFromEligibility(ctx, req.UserID, l, resp); persistErr != nil {
		s.logger.Warn("failed to persist eligibility", baseLogger.Module("profile"), zap.String("userId", req.UserID), zap.Error(persistErr))
	}

	go s.syncUpstream(context.Background(), req.UserID, resp.AvailableLimit, false, false, deriveStatusFromEligibility(resp))

	return resp, nil
}

// GetCustomerStatus retrieves customer basic eligibility
func (s *profileService) GetCustomerStatus(ctx context.Context, req req.CustomerStatusRequest) (*res.CustomerStatusResponse, error) {
	contact, err := s.contactResolver.Resolve(ctx, req.UserID)
	if err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeUserContactNotFound, 422, "Unable to resolve user contact details (mobile required)")
	}

	resp, err := s.gateway.GetCustomerStatus(ctx, contact.Mobile, contact.Email)
	if err != nil {
		return nil, err
	}
	resp.UserID = req.UserID

	l := lender.Lazypay.String()
	if persistErr := s.repo.UpsertFromCustomerStatus(ctx, req.UserID, l, resp); persistErr != nil {
		s.logger.Warn("failed to persist customer status", baseLogger.Module("profile"), zap.String("userId", req.UserID), zap.Error(persistErr))
	}

	go s.syncUpstream(context.Background(), req.UserID, resp.AvailableLimit, resp.PreApproved, !resp.OnboardingRequired, deriveStatusFromCustomerStatus(resp))

	return resp, nil
}

// GetCombinedProfile fetches combined profile (optionally with txn eligibility)
func (s *profileService) GetCombinedProfile(ctx context.Context, userID string, amount *float64, currency, source string) (*res.UserProfileResponse, error) {
	contact, err := s.contactResolver.Resolve(ctx, userID)
	if err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeUserContactNotFound, 422, "Unable to resolve user contact details (mobile required)")
	}

	var (
		csResp   *res.CustomerStatusResponse
		eligResp *res.EligibilityResponse
		csErr    error
		eligErr  error
	)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		csResp, csErr = s.gateway.GetCustomerStatus(gCtx, contact.Mobile, contact.Email)
		return csErr
	})

	if amount != nil && *amount > 0 {
		g.Go(func() error {
			eligResp, eligErr = s.gateway.CheckEligibility(gCtx, contact.Mobile, contact.Email, *amount)
			return eligErr
		})
	}

	if err := g.Wait(); err != nil {
		if csErr != nil {
			return nil, csErr
		}
		if eligErr != nil {
			s.logger.Warn("eligibility call failed in combined profile", baseLogger.Module("profile"), zap.String("userId", userID), zap.Error(eligErr))
		}
	}

	combined := mergeToUserProfile(csResp, eligResp, userID)

	l := lender.Lazypay.String()
	if persistErr := s.repo.UpsertFromCombined(ctx, userID, l, combined); persistErr != nil {
		s.logger.Warn("failed to persist combined profile", baseLogger.Module("profile"), zap.String("userId", userID), zap.Error(persistErr))
	}

	go s.syncUpstream(context.Background(), userID, combined.AvailableLimit, combined.PreApproved, !combined.OnboardingRequired, combined.Status)

	return combined, nil
}

func (s *profileService) syncUpstream(ctx context.Context, userID string, limit float64, preApproved, onboardingDone bool, status string) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("panic in syncUpstream recovered", baseLogger.Module("profile"), zap.String("userId", userID))
		}
	}()
	if s.profileClient == nil {
		return
	}
	if err := s.profileClient.UpdateLenderProfile(ctx, userprofile.LenderProfileUpdateRequest{
		UserID:         userID,
		Lender:         lender.Lazypay.String(),
		AvailableLimit: limit,
		PreApproved:    preApproved,
		OnboardingDone: onboardingDone,
		CurrentStatus:  status,
	}); err != nil {
		s.logger.Warn("upstream profile sync failed", baseLogger.Module("profile"), zap.String("userId", userID), zap.Error(err))
	}
}

func mergeToUserProfile(cs *res.CustomerStatusResponse, elig *res.EligibilityResponse, userID string) *res.UserProfileResponse {
	resp := &res.UserProfileResponse{
		UserID:               userID,
		Provider:             lender.Lazypay.String(),
		PreApproved:          cs.PreApproved,
		OnboardingRequired:   cs.OnboardingRequired,
		CustomerInfoRequired: cs.CustomerInfoRequired,
		NTBEligible:          cs.NTBEligible,
		AvailableLimit:       cs.AvailableLimit,
		Status:               deriveStatusFromCS(cs),
	}
	if elig != nil {
		resp.TxnEligible = &elig.TxnEligible
		resp.EmiPlans = elig.EmiPlans
		eu := elig.ExistingUser
		resp.ExistingUser = &eu
		resp.ReasonCode = elig.ReasonCode
		resp.ReasonMessage = elig.ReasonMessage
		if elig.AvailableLimit > 0 {
			resp.AvailableLimit = elig.AvailableLimit
		}
		resp.Status = deriveStatusFromCSAndElig(cs, elig)
	}
	return resp
}

func deriveStatusFromEligibility(elig *res.EligibilityResponse) string {
	if elig.TxnEligible && elig.AvailableLimit > 0 {
		return "ACTIVE"
	}
	return "INELIGIBLE"
}

func deriveStatusFromCustomerStatus(cs *res.CustomerStatusResponse) string {
	if cs.OnboardingRequired {
		if cs.NTBEligible != nil && *cs.NTBEligible {
			return "NTB"
		}
		return "NOT_STARTED"
	}
	if cs.PreApproved && cs.AvailableLimit > 0 {
		return "ACTIVE"
	}
	return "INELIGIBLE"
}

func deriveStatusFromCS(cs *res.CustomerStatusResponse) string {
	return deriveStatusFromCustomerStatus(cs)
}

func deriveStatusFromCSAndElig(cs *res.CustomerStatusResponse, elig *res.EligibilityResponse) string {
	if cs.OnboardingRequired {
		if cs.NTBEligible != nil && *cs.NTBEligible {
			return "NTB"
		}
		return "NOT_STARTED"
	}
	if elig.TxnEligible && elig.AvailableLimit > 0 {
		return "ACTIVE"
	}
	return "INELIGIBLE"
}
