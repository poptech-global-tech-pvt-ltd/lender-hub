package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	infra "lending-hub-service/internal/infrastructure/postgres"
	"lending-hub-service/internal/domain/order/entity"
	"lending-hub-service/internal/domain/order/port"
)

// postgresOrderRepository implements OrderRepository using GORM
type postgresOrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new Postgres-backed OrderRepository
func NewOrderRepository(db *gorm.DB) port.OrderRepository {
	return &postgresOrderRepository{db: db}
}

// Create creates a new order record
func (r *postgresOrderRepository) Create(ctx context.Context, order *entity.Order) error {
	model := toOrderModel(order)
	return r.db.WithContext(ctx).Create(&model).Error
}

// GetByPaymentID retrieves an order by payment_id
func (r *postgresOrderRepository) GetByPaymentID(ctx context.Context, paymentID string) (*entity.Order, error) {
	var model infra.LenderPaymentState
	err := r.db.WithContext(ctx).
		Where("payment_id = ?", paymentID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toOrderEntity(&model), nil
}

// GetByLoanID retrieves an order by lender_merchant_txn_id (loanId = lps_xxx)
func (r *postgresOrderRepository) GetByLoanID(ctx context.Context, loanID string) (*entity.Order, error) {
	var model infra.LenderPaymentState
	err := r.db.WithContext(ctx).
		Where("lender_merchant_txn_id = ?", loanID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toOrderEntity(&model), nil
}

// GetByLenderOrderID retrieves an order by lender_order_id (Lazypay's orderId) — recon, no enquiry
func (r *postgresOrderRepository) GetByLenderOrderID(ctx context.Context, lenderOrderID string) (*entity.Order, error) {
	var model infra.LenderPaymentState
	err := r.db.WithContext(ctx).
		Where("lender_order_id = ?", lenderOrderID).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toOrderEntity(&model), nil
}

// GetForUpdate retrieves an order with row lock for update
func (r *postgresOrderRepository) GetForUpdate(ctx context.Context, paymentID string) (*entity.Order, error) {
	var model infra.LenderPaymentState
	err := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("payment_id = ?", paymentID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toOrderEntity(&model), nil
}

// Update updates an existing order record
func (r *postgresOrderRepository) Update(ctx context.Context, order *entity.Order) error {
	model := toOrderModel(order)
	return r.db.WithContext(ctx).
		Where("payment_id = ?", order.PaymentID).
		Updates(&model).Error
}

// ListByUserID lists orders by user with optional merchant and status filters
func (r *postgresOrderRepository) ListByUserID(ctx context.Context, userID, merchantID, status string, page, perPage int) ([]*entity.Order, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	q := r.db.WithContext(ctx).Model(&infra.LenderPaymentState{}).Where("user_id = ?", userID)
	if merchantID != "" {
		q = q.Where("merchant_id = ?", merchantID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []infra.LenderPaymentState
	err := q.Order("created_at DESC").Limit(perPage).Offset(offset).Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	orders := make([]*entity.Order, len(models))
	for i := range models {
		orders[i] = toOrderEntity(&models[i])
	}
	return orders, int(total), nil
}

// postgresPaymentMappingRepository implements PaymentMappingRepository using GORM
type postgresPaymentMappingRepository struct {
	db *gorm.DB
}

// NewPaymentMappingRepository creates a new Postgres-backed PaymentMappingRepository
func NewPaymentMappingRepository(db *gorm.DB) port.PaymentMappingRepository {
	return &postgresPaymentMappingRepository{db: db}
}

// Create creates a new payment mapping record
func (r *postgresPaymentMappingRepository) Create(ctx context.Context, mapping *entity.PaymentMapping) error {
	model := toMappingModel(mapping)
	return r.db.WithContext(ctx).Create(&model).Error
}

// GetByMerchantTxnID retrieves a payment mapping by lender_merchant_txn_id
func (r *postgresPaymentMappingRepository) GetByMerchantTxnID(ctx context.Context, lenderMerchantTxnID string) (*entity.PaymentMapping, error) {
	var model infra.LenderPaymentMapping
	err := r.db.WithContext(ctx).
		Where("lender_merchant_txn_id = ?", lenderMerchantTxnID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toMappingEntity(&model), nil
}

// GetByPaymentID retrieves a payment mapping by payment_id (caller's idempotency key)
func (r *postgresPaymentMappingRepository) GetByPaymentID(ctx context.Context, paymentID string) (*entity.PaymentMapping, error) {
	var model infra.LenderPaymentMapping
	err := r.db.WithContext(ctx).
		Where("payment_id = ?", paymentID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return toMappingEntity(&model), nil
}

// toOrderEntity converts GORM model to domain entity
func toOrderEntity(model *infra.LenderPaymentState) *entity.Order {
	status := entity.OrderStatus(model.Status)
	if status == "" || status == "NULL" || status == "null" {
		status = entity.OrderPending
	}
	lenderLastStatus := model.LenderLastStatus
	if lenderLastStatus != nil && (*lenderLastStatus == "NULL" || *lenderLastStatus == "" || *lenderLastStatus == "null") {
		lenderLastStatus = nil
	}
	return &entity.Order{
		ID:                   model.ID,
		PaymentID:            model.PaymentID,
		UserID:               model.UserID,
		MerchantID:           model.MerchantID,
		Lender:               model.Lender,
		Amount:               model.Amount,
		Currency:             model.Currency,
		Status:               status,
		Source:               model.Source,
		ReturnURL:            model.ReturnURL,
		EMIPlan:              model.EMIPlan,
		LenderOrderID:        model.LenderOrderID,
		LenderMerchantTxnID:  model.LenderMerchantTxnID,
		LenderLastStatus:     lenderLastStatus,
		LenderLastTxnID:      model.LenderLastTxnID,
		LenderLastTxnStatus:  model.LenderLastTxnStatus,
		LenderLastTxnMessage: model.LenderLastTxnMessage,
		LenderLastTxnTime:    model.LenderLastTxnTime,
		LastErrorCode:        model.LastErrorCode,
		LastErrorMessage:     model.LastErrorMessage,
		CreatedAt:            model.CreatedAt,
		UpdatedAt:            model.UpdatedAt,
	}
}

// toOrderModel converts domain entity to GORM model
func toOrderModel(e *entity.Order) *infra.LenderPaymentState {
	// Normalize status: empty, "NULL", or invalid values → PENDING; COMPLETE → SUCCESS
	status := entity.OrderStatus(e.Status).OrDefault().NormalizeForDB()
	return &infra.LenderPaymentState{
		ID:                   e.ID,
		PaymentID:            e.PaymentID,
		UserID:               e.UserID,
		MerchantID:           e.MerchantID,
		Lender:               e.Lender,
		Amount:               e.Amount,
		Currency:             e.Currency,
		Status:               string(status),
		Source:               e.Source,
		ReturnURL:            e.ReturnURL,
		EMIPlan:              e.EMIPlan,
		LenderOrderID:        e.LenderOrderID,
		LenderMerchantTxnID:  e.LenderMerchantTxnID,
		LenderLastStatus:     e.LenderLastStatus,
		LenderLastTxnID:      e.LenderLastTxnID,
		LenderLastTxnStatus:  e.LenderLastTxnStatus,
		LenderLastTxnMessage: e.LenderLastTxnMessage,
		LenderLastTxnTime:    e.LenderLastTxnTime,
		LastErrorCode:        e.LastErrorCode,
		LastErrorMessage:     e.LastErrorMessage,
		CreatedAt:            e.CreatedAt,
		UpdatedAt:            e.UpdatedAt,
	}
}

// toMappingEntity converts GORM model to domain entity
func toMappingEntity(model *infra.LenderPaymentMapping) *entity.PaymentMapping {
	return &entity.PaymentMapping{
		ID:                   model.ID,
		PaymentID:            model.PaymentID,
		UserID:               model.UserID,
		Lender:               model.Lender,
		LenderMerchantTxnID:  model.LenderMerchantTxnID,
		LenderOrderID:        model.LenderOrderID,
		EligibilityResponseID: model.EligibilityResponseID,
		CreatedAt:            model.CreatedAt,
		UpdatedAt:            model.UpdatedAt,
	}
}

// toMappingModel converts domain entity to GORM model
func toMappingModel(e *entity.PaymentMapping) *infra.LenderPaymentMapping {
	return &infra.LenderPaymentMapping{
		ID:                   e.ID,
		PaymentID:            e.PaymentID,
		UserID:               e.UserID,
		Lender:               e.Lender,
		LenderMerchantTxnID:  e.LenderMerchantTxnID,
		LenderOrderID:        e.LenderOrderID,
		EligibilityResponseID: e.EligibilityResponseID,
		CreatedAt:            e.CreatedAt,
		UpdatedAt:            e.UpdatedAt,
	}
}
