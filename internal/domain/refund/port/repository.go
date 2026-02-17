package port

import (
	"context"

	"lending-hub-service/internal/domain/refund/entity"
)

// RefundRepository manages refund state
type RefundRepository interface {
	Create(ctx context.Context, refund *entity.Refund) error
	GetByRefundID(ctx context.Context, refundID string) (*entity.Refund, error)
	GetByProviderRefID(ctx context.Context, lender, providerRefID string) (*entity.Refund, error)
	ListByPaymentID(ctx context.Context, paymentID string) ([]*entity.Refund, error)
	Update(ctx context.Context, refund *entity.Refund) error
}
