package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	req "lending-hub-service/internal/domain/onboarding/dto/request"
	res "lending-hub-service/internal/domain/onboarding/dto/response"
	"lending-hub-service/internal/domain/onboarding/entity"
	"lending-hub-service/internal/domain/onboarding/port"
	profileService "lending-hub-service/internal/domain/profile/service"
	sharedErrors "lending-hub-service/internal/shared/errors"
)

// OnboardingService handles onboarding operations
type OnboardingService struct {
	repo           port.OnboardingRepository
	eventStore     port.OnboardingEventStore
	gateway        port.OnboardingGateway
	profileUpdater *profileService.ProfileUpdater
	processor      *EventProcessor
}

// NewOnboardingService creates a new OnboardingService
func NewOnboardingService(
	repo port.OnboardingRepository,
	eventStore port.OnboardingEventStore,
	gateway port.OnboardingGateway,
	profileUpdater *profileService.ProfileUpdater,
) *OnboardingService {
	return &OnboardingService{
		repo:           repo,
		eventStore:     eventStore,
		gateway:        gateway,
		profileUpdater: profileUpdater,
		processor:      NewEventProcessor(),
	}
}

// StartOnboarding initiates a new onboarding flow
func (s *OnboardingService) StartOnboarding(ctx context.Context, req req.StartOnboardingRequest) (*res.OnboardingResponse, error) {
	// Generate onboarding ID
	onboardingID := uuid.NewString()

	// Call gateway to start onboarding
	gatewayResp, err := s.gateway.StartOnboarding(ctx, req)
	if err != nil {
		return nil, err
	}

	// Create onboarding entity
	onboarding := &entity.Onboarding{
		OnboardingID:         onboardingID,
		ProviderOnboardingID: &gatewayResp.OnboardingID,
		UserID:               req.UserID,
		MerchantID:           req.MerchantID,
		Provider:             "LAZYPAY", // TODO: make configurable
		Mobile:               req.UserContact.Mobile,
		Source:               req.Source,
		Channel:              &req.ChannelID,
		Status:               entity.OnboardingPending,
		IsRetryable:          false,
		RetryCount:           0,
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
	}

	// Serialize raw request
	rawRequest, _ := json.Marshal(req)
	onboarding.RawRequest = rawRequest

	// Serialize raw response
	rawResponse, _ := json.Marshal(gatewayResp)
	onboarding.RawResponse = rawResponse

	// Persist onboarding
	if err := s.repo.Create(ctx, onboarding); err != nil {
		return nil, err
	}

	return &res.OnboardingResponse{
		OnboardingID:    onboardingID,
		OnboardingTxnID: req.OnboardingTxnID,
		Provider:        onboarding.Provider,
		RedirectURL:     gatewayResp.RedirectURL,
		Status:          string(entity.OnboardingPending),
	}, nil
}

// GetStatus retrieves onboarding status
func (s *OnboardingService) GetStatus(ctx context.Context, userID, onboardingID, merchantID string) (*res.OnboardingStatusResponse, error) {
	var onboarding *entity.Onboarding
	var err error

	if onboardingID != "" {
		onboarding, err = s.repo.GetByOnboardingID(ctx, onboardingID)
	} else {
		// Get latest onboarding for user+merchant
		// For now, we'll need to add a method to get latest, or use GetByOnboardingID with a pattern
		// This is a simplified version - in production you'd have GetLatestByUserAndMerchant
		return nil, sharedErrors.New(sharedErrors.CodeOnboardingNotFound, 404, "onboarding not found")
	}

	if err != nil {
		return nil, err
	}
	if onboarding == nil {
		return nil, sharedErrors.New(sharedErrors.CodeOnboardingNotFound, 404, "onboarding not found")
	}

	// Validate user and merchant match
	if onboarding.UserID != userID || onboarding.MerchantID != merchantID {
		return nil, sharedErrors.New(sharedErrors.CodeOnboardingNotFound, 404, "onboarding not found")
	}

	// Fetch events
	events, err := s.eventStore.ListByOnboardingID(ctx, onboarding.OnboardingID)
	if err != nil {
		return nil, err
	}

	// Build steps from events
	steps := s.buildStepsFromEvents(events)

	// Build response
	response := &res.OnboardingStatusResponse{
		OnboardingID:           onboarding.OnboardingID,
		UserID:                 onboarding.UserID,
		Provider:               onboarding.Provider,
		Status:                 string(onboarding.Status),
		COFEligible:            onboarding.COFEligible,
		RejectionReasonCode:    onboarding.RejectionReasonCode,
		RejectionReasonMessage: onboarding.RejectionReasonMessage,
		Retrying:               onboarding.IsRetryable && onboarding.NextRetryAt != nil && onboarding.NextRetryAt.After(time.Now()),
		RetryCount:             onboarding.RetryCount,
		UpdatedAt:              onboarding.UpdatedAt.Format(time.RFC3339),
		Steps:                  steps,
	}

	if onboarding.LastStep != nil {
		lastStepStr := string(*onboarding.LastStep)
		response.LastStep = &lastStepStr
	}

	if onboarding.NextRetryAt != nil {
		nextRetryStr := onboarding.NextRetryAt.Format(time.RFC3339)
		response.NextRetryAt = &nextRetryStr
	}

	return response, nil
}

// ProcessCallback processes an onboarding callback event
func (s *OnboardingService) ProcessCallback(ctx context.Context, req req.OnboardingCallbackRequest) error {
	// Parse event time
	eventTime, err := time.Parse(time.RFC3339, req.EventTime)
	if err != nil {
		return sharedErrors.New(sharedErrors.CodeInvalidRequest, 400, "invalid eventTime format")
	}

	// Build event entity
	var step *entity.OnboardingStep
	if req.Step != nil {
		stepVal := entity.OnboardingStep(*req.Step)
		step = &stepVal
	}

	event := &entity.OnboardingEvent{
		Provider:     req.Provider,
		MerchantID:   req.MerchantID,
		UserID:       req.UserID,
		Mobile:       req.Mobile,
		OnboardingID: req.OnboardingID,
		EventType:    req.EventType,
		Status:       req.Status,
		Step:         step,
		ErrorCode:    req.ErrorCode,
		Message:      req.Message,
		EventTime:    eventTime,
		CreatedAt:    time.Now().UTC(),
	}

	// Serialize raw payload
	rawPayload, _ := json.Marshal(req)
	event.RawPayload = rawPayload

	// Append to event store (idempotent via unique constraint)
	err = s.eventStore.Append(ctx, event)
	if err != nil {
		// If unique constraint violation, ignore (idempotent)
		// In production, you'd check for specific error type
		// For now, we'll log and continue
	}

	// Get onboarding for update
	onboarding, err := s.repo.GetForUpdate(ctx, req.OnboardingID)
	if err != nil {
		return err
	}
	if onboarding == nil {
		return sharedErrors.New(sharedErrors.CodeOnboardingNotFound, 404, "onboarding not found")
	}

	// Process event to determine new status
	newStatus := s.processor.ProcessEvent(event, onboarding.Status)
	onboarding.Status = newStatus
	onboarding.LastStep = step

	// Classify error if present
	if event.ErrorCode != nil {
		classification, found := ClassifyError(*event.ErrorCode)
		if found {
			onboarding.RejectionReasonCode = &classification.CanonicalCode
			if event.Message != nil {
				onboarding.RejectionReasonMessage = event.Message
			} else {
				msg := classification.UserMessage
				onboarding.RejectionReasonMessage = &msg
			}
			onboarding.IsRetryable = classification.IsRetryable

			// Calculate next retry time if retryable
			if classification.IsRetryable {
				nextRetry := s.processor.CalculateNextRetryTime(event.ErrorCode, onboarding.RetryCount)
				onboarding.NextRetryAt = nextRetry
				if nextRetry != nil {
					now := time.Now().UTC()
					onboarding.LastRetryAt = &now
					onboarding.RetryCount++
				}
			}
		}
	}

	// Update onboarding
	if err := s.repo.Update(ctx, onboarding); err != nil {
		return err
	}

	// If SUCCESS, update profile
	if newStatus == entity.OnboardingSuccess {
		// Get credit limit from gateway response or use default
		creditLimit := 50000.0 // TODO: get from gateway response
		if err := s.profileUpdater.UpdateOnOnboardingSuccess(ctx, onboarding.UserID, onboarding.Provider, creditLimit); err != nil {
			// Log error but don't fail callback
			// In production, you might want to retry this
		}
	}

	return nil
}

// buildStepsFromEvents builds step details from events
func (s *OnboardingService) buildStepsFromEvents(events []*entity.OnboardingEvent) []res.StepDetail {
	stepMap := make(map[entity.OnboardingStep]*res.StepDetail)
	allSteps := []entity.OnboardingStep{
		entity.StepUserData,
		entity.StepEMISelection,
		entity.StepKYC,
		entity.StepKFS,
		entity.StepMITC,
		entity.StepAutopay,
	}

	// Initialize all steps as PENDING
	for _, step := range allSteps {
		stepStr := string(step)
		stepMap[step] = &res.StepDetail{
			Step:   stepStr,
			Status: string(entity.StepPending),
		}
	}

	// Process events to update step statuses
	for _, event := range events {
		if event.Step == nil {
			continue
		}
		stepDetail, exists := stepMap[*event.Step]
		if !exists {
			continue
		}

		switch event.Status {
		case "SUCCESS":
			stepDetail.Status = string(entity.StepSuccess)
			completedAt := event.EventTime.Format(time.RFC3339)
			stepDetail.CompletedAt = &completedAt
		case "FAILED":
			stepDetail.Status = string(entity.StepFailed)
		case "IN_PROGRESS":
			stepDetail.Status = string(entity.StepPending)
		}
	}

	// Convert map to slice
	steps := make([]res.StepDetail, 0, len(allSteps))
	for _, step := range allSteps {
		steps = append(steps, *stepMap[step])
	}

	return steps
}
