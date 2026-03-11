package repository

import (
	"context"

	"gorm.io/gorm"

	infra "lending-hub-service/internal/infrastructure/postgres"
	"lending-hub-service/internal/domain/onboarding/entity"
	"lending-hub-service/internal/domain/onboarding/port"
)

// postgresOnboardingEventStore implements OnboardingEventStore using GORM
type postgresOnboardingEventStore struct {
	db *gorm.DB
}

// NewOnboardingEventStore creates a new Postgres-backed OnboardingEventStore
func NewOnboardingEventStore(db *gorm.DB) port.OnboardingEventStore {
	return &postgresOnboardingEventStore{db: db}
}

// Append appends a new event to the event log
func (s *postgresOnboardingEventStore) Append(ctx context.Context, event *entity.OnboardingEvent) error {
	model := toEventModel(event)
	return s.db.WithContext(ctx).Create(&model).Error
}

// ListByOnboardingID lists all events for an onboarding, ordered by event_time ascending
func (s *postgresOnboardingEventStore) ListByOnboardingID(ctx context.Context, onboardingID string) ([]*entity.OnboardingEvent, error) {
	var models []infra.LenderOnboardingEvent
	err := s.db.WithContext(ctx).
		Where("onboarding_id = ?", onboardingID).
		Order("event_time ASC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	events := make([]*entity.OnboardingEvent, len(models))
	for i, model := range models {
		events[i] = toEventEntity(&model)
	}

	return events, nil
}

// toEventEntity converts GORM model to domain entity
func toEventEntity(model *infra.LenderOnboardingEvent) *entity.OnboardingEvent {
	var step *entity.OnboardingStep
	if model.Step != nil {
		stepVal := entity.OnboardingStep(*model.Step)
		step = &stepVal
	}

	return &entity.OnboardingEvent{
		ID:           model.ID,
		Provider:     model.Provider,
		MerchantID:   model.MerchantID,
		UserID:       model.UserID,
		Mobile:       model.Mobile,
		OnboardingID: model.OnboardingID,
		EventType:    model.EventType,
		Status:       model.Status,
		Step:         step,
		ErrorCode:    model.ErrorCode,
		Message:      model.Message,
		EventTime:    model.EventTime,
		RawPayload:   model.RawPayload,
		CreatedAt:    model.CreatedAt,
	}
}

// toEventModel converts domain entity to GORM model
func toEventModel(e *entity.OnboardingEvent) *infra.LenderOnboardingEvent {
	var step *string
	if e.Step != nil {
		stepStr := string(*e.Step)
		step = &stepStr
	}

	return &infra.LenderOnboardingEvent{
		ID:           e.ID,
		Provider:     e.Provider,
		MerchantID:   e.MerchantID,
		UserID:       e.UserID,
		Mobile:       e.Mobile,
		OnboardingID: e.OnboardingID,
		EventType:    e.EventType,
		Status:       e.Status,
		Step:         step,
		ErrorCode:    e.ErrorCode,
		Message:      e.Message,
		EventTime:    e.EventTime,
		RawPayload:   e.RawPayload,
		CreatedAt:    e.CreatedAt,
	}
}
