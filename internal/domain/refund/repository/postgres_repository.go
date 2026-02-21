package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

// ListByLoanID lists all refunds for an order by loanId
func (r *postgresRefundRepository) ListByLoanID(ctx context.Context, loanID string) ([]*entity.Refund, error) {
	var models []infra.LenderRefund
	err := r.db.WithContext(ctx).
		Where("loan_id = ?", loanID).
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

// GetByPaymentRefundID retrieves a refund by lender and payment_refund_id
func (r *postgresRefundRepository) GetByPaymentRefundID(ctx context.Context, lender, paymentRefundID string) (*entity.Refund, error) {
	var model infra.LenderRefund
	err := r.db.WithContext(ctx).
		Where("lender = ? AND payment_refund_id = ?", lender, paymentRefundID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toEntity(&model), nil
}

// GetForUpdateByPaymentRefundID retrieves a refund by lender and payment_refund_id with row lock
func (r *postgresRefundRepository) GetForUpdateByPaymentRefundID(ctx context.Context, lender, paymentRefundID string) (*entity.Refund, error) {
	var model infra.LenderRefund
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("lender = ? AND payment_refund_id = ?", lender, paymentRefundID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toEntity(&model), nil
}

// GetForUpdateByRefundID retrieves a refund by refund_id with row lock
func (r *postgresRefundRepository) GetForUpdateByRefundID(ctx context.Context, refundID string) (*entity.Refund, error) {
	var model infra.LenderRefund
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
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

// ListByUserID lists refunds for a user with pagination
func (r *postgresRefundRepository) ListByUserID(ctx context.Context, userID string, page, perPage int) ([]*entity.Refund, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var total int64
	if err := r.db.WithContext(ctx).Model(&infra.LenderRefund{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []infra.LenderRefund
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(perPage).
		Offset(offset).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	refunds := make([]*entity.Refund, len(models))
	for i, model := range models {
		refunds[i] = toEntity(&model)
	}
	return refunds, int(total), nil
}

// toEntity converts GORM model to domain entity
func toEntity(model *infra.LenderRefund) *entity.Refund {
	refund := &entity.Refund{
		ID:                     model.ID,
		RefundID:               model.RefundID,
		PaymentRefundID:        model.PaymentRefundID,
		PaymentID:              model.PaymentID,
		LoanID:                 model.LoanID,
		UserID:                 model.UserID,
		Lender:                 model.Lender,
		Amount:                 model.Amount,
		Currency:               model.Currency,
		Status:                 entity.RefundStatus(model.Status),
		ProviderMerchantTxnID:  model.ProviderMerchantTxnID,
		ProviderParentTxnID:    model.ProviderParentTxnID,
		ProviderRefundTxnID:    model.ProviderRefundTxnID,
		ProviderRefundRefID:    model.ProviderRefundRefID,
		LenderRefID:            model.LenderRefID,
		LenderStatus:           model.LenderStatus,
		LenderMessage:          model.LenderMessage,
		LastEnquiredAt:         model.LastEnquiredAt,
		CreatedAt:              model.CreatedAt,
		UpdatedAt:              model.UpdatedAt,
	}

	// Map reason
	if model.Reason != nil {
		reason := entity.RefundReason(*model.Reason)
		refund.Reason = &reason
	}

	return refund
}

// toModel converts domain entity to GORM model
func toModel(e *entity.Refund) *infra.LenderRefund {
	var reason *string
	if e.Reason != nil {
		reasonStr := string(*e.Reason)
		reason = &reasonStr
	}

	return &infra.LenderRefund{
		ID:                     e.ID,
		RefundID:               e.RefundID,
		PaymentRefundID:        e.PaymentRefundID,
		PaymentID:              e.PaymentID,
		LoanID:                 e.LoanID,
		UserID:                 e.UserID,
		Lender:                 e.Lender,
		Amount:                 e.Amount,
		Currency:               e.Currency,
		Status:                 string(e.Status),
		Reason:                 reason,
		ProviderMerchantTxnID:  e.ProviderMerchantTxnID,
		ProviderParentTxnID:    e.ProviderParentTxnID,
		ProviderRefundTxnID:    e.ProviderRefundTxnID,
		ProviderRefundRefID:    e.ProviderRefundRefID,
		LenderRefID:            e.LenderRefID,
		LenderStatus:           e.LenderStatus,
		LenderMessage:          e.LenderMessage,
		LastEnquiredAt:         e.LastEnquiredAt,
		CreatedAt:              e.CreatedAt,
		UpdatedAt:              e.UpdatedAt,
	}
}
