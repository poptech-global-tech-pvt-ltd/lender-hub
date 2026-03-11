package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	infra "lending-hub-service/internal/infrastructure/postgres"
	"lending-hub-service/internal/domain/order/entity"
	"lending-hub-service/internal/domain/order/port"
)

// postgresIdempotencyRepository implements IdempotencyRepository using GORM
type postgresIdempotencyRepository struct {
	db *gorm.DB
}

// NewIdempotencyRepository creates a new Postgres-backed IdempotencyRepository
func NewIdempotencyRepository(db *gorm.DB) port.IdempotencyRepository {
	return &postgresIdempotencyRepository{db: db}
}

// TryAcquire attempts to create a PROCESSING record; returns existing record if conflict
func (r *postgresIdempotencyRepository) TryAcquire(ctx context.Context, key, requestHash string) (*entity.IdempotencyKey, error) {
	// First, try to get existing record
	var existing infra.LenderIdempotencyKey
	err := r.db.WithContext(ctx).
		Where("idempotency_key = ?", key).
		First(&existing).Error

	if err == nil {
		// Record exists, return it
		return toIdempotencyEntity(&existing), nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Unexpected error
		return nil, err
	}

	// Record doesn't exist, try to create new one with PROCESSING status
	// Note: This is a simplified version. In production, you'd want to handle
	// unique constraint violations more carefully, possibly using database-level
	// transactions or SELECT FOR UPDATE.
	now := time.Now().UTC()
	newRecord := &infra.LenderIdempotencyKey{
		IdempotencyKey:  key,
		RequestHash:     requestHash,
		Status:          "PROCESSING",
		ResponsePayload: nil,
		LenderOrderID:   nil,
		CreatedAt:       now,
		ExpiresAt:       now.Add(24 * time.Hour), // Expire after 24 hours
	}

	createErr := r.db.WithContext(ctx).Create(newRecord).Error
	if createErr != nil {
		// If unique constraint violation, fetch the existing record
		if isUniqueConstraintError(createErr) {
			var existingRecord infra.LenderIdempotencyKey
			fetchErr := r.db.WithContext(ctx).
				Where("idempotency_key = ?", key).
				First(&existingRecord).Error
			if fetchErr != nil {
				return nil, fetchErr
			}
			return toIdempotencyEntity(&existingRecord), nil
		}
		return nil, createErr
	}

	return toIdempotencyEntity(newRecord), nil
}

// MarkCompleted updates status to COMPLETED and stores response payload
func (r *postgresIdempotencyRepository) MarkCompleted(ctx context.Context, key string, responsePayload []byte, lenderOrderID *string) error {
	updates := map[string]interface{}{
		"status":           "COMPLETED",
		"response_payload": responsePayload,
	}
	if lenderOrderID != nil {
		updates["lender_order_id"] = *lenderOrderID
	}

	return r.db.WithContext(ctx).
		Model(&infra.LenderIdempotencyKey{}).
		Where("idempotency_key = ?", key).
		Updates(updates).Error
}

// MarkFailed updates status to FAILED
func (r *postgresIdempotencyRepository) MarkFailed(ctx context.Context, key string) error {
	return r.db.WithContext(ctx).
		Model(&infra.LenderIdempotencyKey{}).
		Where("idempotency_key = ?", key).
		Update("status", "FAILED").Error
}

// Get returns current record for given key
func (r *postgresIdempotencyRepository) Get(ctx context.Context, key string) (*entity.IdempotencyKey, error) {
	var model infra.LenderIdempotencyKey
	err := r.db.WithContext(ctx).
		Where("idempotency_key = ?", key).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toIdempotencyEntity(&model), nil
}

// toIdempotencyEntity converts GORM model to domain entity
func toIdempotencyEntity(model *infra.LenderIdempotencyKey) *entity.IdempotencyKey {
	return &entity.IdempotencyKey{
		ID:              model.ID,
		Key:             model.IdempotencyKey,
		RequestHash:     model.RequestHash,
		Status:          entity.IdempotencyStatus(model.Status),
		ResponsePayload: model.ResponsePayload,
		LenderOrderID:   model.LenderOrderID,
		CreatedAt:       model.CreatedAt,
		ExpiresAt:       model.ExpiresAt,
	}
}

// isUniqueConstraintError checks if error is a unique constraint violation
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	// Check for PostgreSQL unique constraint error
	// This is a simplified check; in production you might want more robust error handling
	// Check for GORM duplicate key error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	// Check for SQL unique constraint violation (PostgreSQL error code 23505)
	// In a production system, you'd want to parse the error more carefully
	return false
}
