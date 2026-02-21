package service

import (
	"context"
	"time"

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
	GetCombinedProfile(ctx context.Context, userID string, amount float64, currency, source string) (*res.UserProfileResponse, error)
}

type profileService struct {
	gateway         port.ProfileGateway
	repo            port.ProfileRepository
	contactResolver port.ContactResolver
	profileSyncer   port.ProfileSyncer
	mc              interface{} // metrics.MetricsClient if needed
	logger          *baseLogger.Logger
}

// NewProfileService creates a new ProfileService
func NewProfileService(
	gw port.ProfileGateway,
	repo port.ProfileRepository,
	contactResolver port.ContactResolver,
	profileSyncer port.ProfileSyncer,
	mc interface{},
	logger *baseLogger.Logger,
) ProfileService {
	return &profileService{
		gateway:         gw,
		repo:            repo,
		contactResolver: contactResolver,
		profileSyncer:   profileSyncer,
		mc:              mc,
		logger:          logger,
	}
}

// CheckEligibility checks txn-level eligibility
func (s *profileService) CheckEligibility(ctx context.Context, req req.EligibilityRequest) (*res.EligibilityResponse, error) {
	contact, err := s.contactResolver.GetContact(ctx, req.UserID)
	if err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeUserContactNotFound, 422, "Unable to resolve user contact details (mobile required)")
	}

	result, err := s.gateway.CheckEligibility(ctx, contact.Mobile, contact.Email, req.Amount)
	if err != nil {
		return nil, err
	}

	l := lender.Lazypay.String()
	if persistErr := s.repo.UpsertFromEligibility(ctx, req.UserID, l, result); persistErr != nil {
		s.logger.Warn("UpsertFromEligibility failed",
			baseLogger.Module("profile"), zap.String("userId", req.UserID), zap.Error(persistErr))
	}

	// Async upstream sync — fire-and-forget
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("panic in upstream sync", baseLogger.Module("profile"), zap.Any("panic", r))
			}
		}()
		_ = s.profileSyncer.UpdateLenderProfile(context.Background(), userprofile.LenderProfileUpdateRequest{
			UserID:         req.UserID,
			Lender:         l,
			AvailableLimit: result.AvailableLimit,
			PreApproved:    false,
			OnboardingDone: false,
			CurrentStatus:  deriveStatusFromEligibility(result),
		})
	}()

	return mapEligibilityResultToResponse(req.UserID, result), nil
}

// GetCustomerStatus retrieves customer basic eligibility
func (s *profileService) GetCustomerStatus(ctx context.Context, req req.CustomerStatusRequest) (*res.CustomerStatusResponse, error) {
	contact, err := s.contactResolver.GetContact(ctx, req.UserID)
	if err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeUserContactNotFound, 422, "Unable to resolve user contact details (mobile required)")
	}

	result, err := s.gateway.GetCustomerStatus(ctx, contact.Mobile, contact.Email)
	if err != nil {
		return nil, err
	}

	l := lender.Lazypay.String()
	if persistErr := s.repo.UpsertFromCustomerStatus(ctx, req.UserID, l, result); persistErr != nil {
		s.logger.Warn("UpsertFromCustomerStatus failed",
			baseLogger.Module("profile"), zap.String("userId", req.UserID), zap.Error(persistErr))
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("panic in upstream sync", baseLogger.Module("profile"), zap.Any("panic", r))
			}
		}()
		_ = s.profileSyncer.UpdateLenderProfile(context.Background(), userprofile.LenderProfileUpdateRequest{
			UserID:         req.UserID,
			Lender:         l,
			AvailableLimit: result.AvailableLimit,
			PreApproved:    result.PreApproved,
			OnboardingDone: !result.OnboardingRequired,
			CurrentStatus:  deriveStatusFromCustomerStatus(result),
		})
	}()

	return mapCustomerStatusResultToResponse(req.UserID, result), nil
}

// GetCombinedProfile fetches combined profile (optionally with txn eligibility when amount > 0)
func (s *profileService) GetCombinedProfile(ctx context.Context, userID string, amount float64, currency, source string) (*res.UserProfileResponse, error) {
	contact, err := s.contactResolver.GetContact(ctx, userID)
	if err != nil {
		return nil, sharedErrors.New(sharedErrors.CodeUserContactNotFound, 422, "Unable to resolve user contact details (mobile required)")
	}

	var csResult *port.CustomerStatusResult
	var elResult *port.EligibilityResult
	var csErr, elErr error

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		csResult, csErr = s.gateway.GetCustomerStatus(gCtx, contact.Mobile, contact.Email)
		return csErr
	})

	if amount > 0 {
		g.Go(func() error {
			elResult, elErr = s.gateway.CheckEligibility(gCtx, contact.Mobile, contact.Email, amount)
			return nil // non-fatal — log and continue with partial data
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	if elErr != nil {
		s.logger.Warn("eligibility call failed in combined profile",
			baseLogger.Module("profile"), zap.String("userId", userID), zap.Error(elErr))
	}

	l := lender.Lazypay.String()
	if persistErr := s.repo.UpsertFromCombined(ctx, userID, l, csResult, elResult); persistErr != nil {
		s.logger.Warn("UpsertFromCombined failed",
			baseLogger.Module("profile"), zap.String("userId", userID), zap.Error(persistErr))
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("panic in upstream sync", baseLogger.Module("profile"), zap.Any("panic", r))
			}
		}()
		limit := csResult.AvailableLimit
		if elResult != nil && elResult.AvailableLimit > 0 {
			limit = elResult.AvailableLimit
		}
		_ = s.profileSyncer.UpdateLenderProfile(context.Background(), userprofile.LenderProfileUpdateRequest{
			UserID:         userID,
			Lender:         l,
			AvailableLimit: limit,
			PreApproved:    csResult.PreApproved,
			OnboardingDone: !csResult.OnboardingRequired,
			CurrentStatus:  deriveStatusFromCombined(csResult, elResult),
		})
	}()

	return mapCombinedToUserProfileResponse(userID, csResult, elResult), nil
}

func mapEligibilityResultToResponse(userID string, r *port.EligibilityResult) *res.EligibilityResponse {
	now := time.Now().UTC()
	resp := &res.EligibilityResponse{
		UserID:            userID,
		Lender:            lender.Lazypay.String(),
		TxnEligible:       r.TxnEligible,
		EligibilityCode:   orEmpty(r.EligibilityCode),
		EligibilityReason: orEmpty(r.EligibilityReason),
		AvailableLimit:    r.AvailableLimit,
		CreditLimit:       r.AvailableLimit,
		ExistingUser:      r.ExistingUser,
		CheckedAt:         now,
	}
	for _, p := range r.EmiPlans {
		resp.EmiPlans = append(resp.EmiPlans, res.EmiPlan{
			Tenure:             p.Tenure,
			EMI:                p.EMI,
			InterestRate:       p.InterestRate,
			Principal:          p.Principal,
			TotalPayableAmount: p.TotalPayableAmount,
			FirstEmiDueDate:    p.FirstEmiDueDate,
			Type:               p.Type,
		})
	}
	return resp
}

func mapCustomerStatusResultToResponse(userID string, r *port.CustomerStatusResult) *res.CustomerStatusResponse {
	return &res.CustomerStatusResponse{
		UserID:             userID,
		Lender:             lender.Lazypay.String(),
		PreApproved:        r.PreApproved,
		OnboardingRequired: r.OnboardingRequired,
		OnboardingDone:     !r.OnboardingRequired,
		NTBEligible:        r.NTBEligible,
		AvailableLimit:     r.AvailableLimit,
		CheckedAt:          time.Now().UTC(),
	}
}

func mapCombinedToUserProfileResponse(userID string, cs *port.CustomerStatusResult, el *port.EligibilityResult) *res.UserProfileResponse {
	status := deriveStatusFromCombined(cs, el)
	if status == "" {
		status = "NOT_STARTED"
	}
	resp := &res.UserProfileResponse{
		UserID:             userID,
		Lender:             lender.Lazypay.String(),
		Status:             status,
		PreApproved:        cs.PreApproved,
		OnboardingRequired: cs.OnboardingRequired,
		OnboardingDone:     !cs.OnboardingRequired,
		NTBEligible:        cs.NTBEligible,
		AvailableLimit:     cs.AvailableLimit,
		CreditLimit:        cs.AvailableLimit,
		LastCheckedAt:      time.Now().UTC(),
	}
	if el != nil {
		resp.TxnEligible = el.TxnEligible
		resp.EligibilityCode = el.EligibilityCode
		resp.EligibilityReason = el.EligibilityReason
		resp.AvailableLimit = el.AvailableLimit
		resp.CreditLimit = el.AvailableLimit
		for _, p := range el.EmiPlans {
			resp.EmiPlans = append(resp.EmiPlans, res.EmiPlan{
				Tenure:             p.Tenure,
				EMI:                p.EMI,
				InterestRate:       p.InterestRate,
				Principal:          p.Principal,
				TotalPayableAmount: p.TotalPayableAmount,
				FirstEmiDueDate:    p.FirstEmiDueDate,
				Type:               p.Type,
			})
		}
	}
	return resp
}

func deriveStatusFromEligibility(el *port.EligibilityResult) string {
	if el.TxnEligible && el.AvailableLimit > 0 {
		return "ACTIVE"
	}
	return "INELIGIBLE"
}

func deriveStatusFromCustomerStatus(cs *port.CustomerStatusResult) string {
	if cs.OnboardingRequired {
		return "NOT_STARTED"
	}
	if cs.PreApproved && cs.AvailableLimit > 0 {
		return "ACTIVE"
	}
	return "INELIGIBLE"
}

func deriveStatusFromCombined(cs *port.CustomerStatusResult, el *port.EligibilityResult) string {
	if cs.OnboardingRequired {
		return "NOT_STARTED"
	}
	if el != nil {
		if el.TxnEligible && el.AvailableLimit > 0 {
			return "ACTIVE"
		}
		return "INELIGIBLE"
	}
	if cs.PreApproved && cs.AvailableLimit > 0 {
		return "ACTIVE"
	}
	return "INELIGIBLE"
}

func orEmpty(s string) string {
	if s == "" {
		return ""
	}
	return s
}
