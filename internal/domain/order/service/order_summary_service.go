package service

import (
	"context"

	refundEntity "lending-hub-service/internal/domain/refund/entity"
	refundPort "lending-hub-service/internal/domain/refund/port"
	res "lending-hub-service/internal/domain/order/dto/response"
	baseLogger "lending-hub-service/pkg/logger"

	"go.uber.org/zap"
)

// OrderSummaryService provides order + refunds summary in one call
type OrderSummaryService struct {
	orderSvc   *OrderService
	refundRepo refundPort.RefundRepository
	logger     *baseLogger.Logger
}

// NewOrderSummaryService creates a new OrderSummaryService
func NewOrderSummaryService(
	orderSvc *OrderService,
	refundRepo refundPort.RefundRepository,
	logger *baseLogger.Logger,
) *OrderSummaryService {
	return &OrderSummaryService{
		orderSvc:   orderSvc,
		refundRepo: refundRepo,
		logger:     logger,
	}
}

// GetSummary returns order, all refunds, and computed financial summary
func (s *OrderSummaryService) GetSummary(ctx context.Context, paymentID string) (*res.OrderSummaryResponse, error) {
	order, err := s.orderSvc.GetByPaymentID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	loanID := ""
	if order.LenderMerchantTxnID != nil {
		loanID = *order.LenderMerchantTxnID
	}

	refunds, err := s.refundRepo.ListByLoanID(ctx, loanID)
	if err != nil {
		s.logger.Warn("failed to load refunds for summary",
			baseLogger.Module("order"), zap.String("loanId", loanID), zap.Error(err))
		refunds = nil
	}

	var totalRefunded float64
	for _, r := range refunds {
		if r.Status == refundEntity.RefundStatusSuccess {
			totalRefunded += r.Amount
		}
	}
	netAmount := order.Amount - totalRefunded
	if netAmount < 0 {
		netAmount = 0
	}

	refundItems := make([]res.RefundSummaryItem, len(refunds))
	for i, r := range refunds {
		providerRefundTxnID := ""
		if r.ProviderRefundTxnID != nil {
			providerRefundTxnID = *r.ProviderRefundTxnID
		}
		reason := ""
		if r.Reason != nil {
			reason = string(*r.Reason)
		}
		lenderStatus := ""
		if r.LenderStatus != nil {
			lenderStatus = *r.LenderStatus
		}
		refundItems[i] = res.RefundSummaryItem{
			RefundID:            r.RefundID,
			PaymentRefundID:     r.PaymentRefundID,
			ProviderRefundTxnID: providerRefundTxnID,
			Status:              string(r.Status.OrDefault()),
			Amount:              r.Amount,
			Currency:            r.Currency,
			Reason:              reason,
			LenderStatus:        lenderStatus,
			CreatedAt:           r.CreatedAt,
			UpdatedAt:           r.UpdatedAt,
		}
	}

	orderResp := s.orderSvc.orderToStatusResponse(order)

	return &res.OrderSummaryResponse{
		Order:   *orderResp,
		Refunds: refundItems,
		Summary: res.OrderFinancialSummary{
			TotalAmount:   order.Amount,
			TotalRefunded: totalRefunded,
			NetAmount:     netAmount,
			RefundCount:   len(refunds),
			FullyRefunded: netAmount <= 0 && len(refunds) > 0,
		},
	}, nil
}
