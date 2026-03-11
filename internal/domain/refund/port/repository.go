package port

import (
	"context"

	"lending-hub-service/internal/domain/refund/entity"
)

// RefundRepository manages refund state
type RefundRepository interface {
	Create(ctx context.Context, refund *entity.Refund) error

	GetByPaymentRefundID(ctx context.Context, lender, paymentRefundID string) (*entity.Refund, error)
	GetForUpdateByPaymentRefundID(ctx context.Context, lender, paymentRefundID string) (*entity.Refund, error)
	GetByRefundID(ctx context.Context, refundID string) (*entity.Refund, error)
	GetForUpdateByRefundID(ctx context.Context, refundID string) (*entity.Refund, error)

	ListByPaymentID(ctx context.Context, paymentID string) ([]*entity.Refund, error)
	ListByLoanID(ctx context.Context, loanID string) ([]*entity.Refund, error)
	ListByUserID(ctx context.Context, userID string, page, perPage int) ([]*entity.Refund, int, error)

	Update(ctx context.Context, refund *entity.Refund) error
}
