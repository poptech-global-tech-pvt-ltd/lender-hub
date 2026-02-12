package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	infra "lending-hub-service/internal/infrastructure/postgres"
	"lending-hub-service/internal/domain/refund/entity"
	"lending-hub-service/internal/domain/refund/port"
)

// postgresRefundRepository implements RefundRepository using GORM
type postgresRefundRepository struct {
	db *gorm.DB
}

// NewRefundRepository creates a new Postgres-backed RefundRepository
func NewRefundRepository(db *gorm.DB) port.RefundRepository {
	return &postgresRefundRepository{db: db}
}

// Create creates a new refund record
func (r *postgresRefundRepository) Create(ctx context.Context, refund *entity.Refund) error {
	model := toModel(refund)
	return r.db.WithContext(ctx).Create(&model).Error
}

// GetByRefundID retrieves a refund by refund_id
func (r *postgresRefundRepository) GetByRefundID(ctx context.Context, refundID string) (*entity.Refund, error) {
	var model infra.LenderRefund
	err := r.db.WithContext(ctx).
		Where("refund_id = ?", refundID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toEntity(&model), nil
}

// ListByPaymentID lists all refunds for a payment
func (r *postgresRefundRepository) ListByPaymentID(ctx context.Context, paymentID string) ([]*entity.Refund, error) {
	var models []infra.LenderRefund
	err := r.db.WithContext(ctx).
		Where("payment_id = ?", paymentID).
		Order("created_at ASC").
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	refunds := make([]*entity.Refund, len(models))
	for i, model := range models {
		refunds[i] = toEntity(&model)
	}

	return refunds, nil
}

// Update updates an existing refund record
func (r *postgresRefundRepository) Update(ctx context.Context, refund *entity.Refund) error {
	model := toModel(refund)
	return r.db.WithContext(ctx).
		Where("refund_id = ?", refund.RefundID).
		Updates(&model).Error
}

// toEntity converts GORM model to domain entity
func toEntity(model *infra.LenderRefund) *entity.Refund {
	return &entity.Refund{
		ID:            model.ID,
		RefundID:      model.RefundID,
		PaymentID:     model.PaymentID,
		UserID:        model.UserID,
		Lender:        model.Lender,
		Amount:        model.Amount,
		Currency:      model.Currency,
		Status:        model.Status,
		Reason:        model.Reason,
		LenderRefID:   model.LenderRefID,
		LenderStatus:  model.LenderStatus,
		LenderMessage: model.LenderMessage,
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
	}
}

// toModel converts domain entity to GORM model
func toModel(e *entity.Refund) *infra.LenderRefund {
	return &infra.LenderRefund{
		ID:            e.ID,
		RefundID:      e.RefundID,
		PaymentID:     e.PaymentID,
		UserID:        e.UserID,
		Lender:        e.Lender,
		Amount:        e.Amount,
		Currency:      e.Currency,
		Status:        e.Status,
		Reason:        e.Reason,
		LenderRefID:   e.LenderRefID,
		LenderStatus:  e.LenderStatus,
		LenderMessage: e.LenderMessage,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}
