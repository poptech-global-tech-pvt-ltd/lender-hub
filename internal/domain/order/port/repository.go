package port

import (
	"context"

	"lending-hub-service/internal/domain/order/entity"
)

// OrderRepository manages payment/order state
type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	GetByPaymentID(ctx context.Context, paymentID string) (*entity.Order, error)
	GetForUpdate(ctx context.Context, paymentID string) (*entity.Order, error)
	Update(ctx context.Context, order *entity.Order) error
}

// PaymentMappingRepository manages lender_payment_mapping
type PaymentMappingRepository interface {
	Create(ctx context.Context, mapping *entity.PaymentMapping) error
	GetByMerchantTxnID(ctx context.Context, lenderMerchantTxnID string) (*entity.PaymentMapping, error)
	GetByPaymentID(ctx context.Context, paymentID string) (*entity.PaymentMapping, error) // Look up merchantTxnId from paymentId
}
