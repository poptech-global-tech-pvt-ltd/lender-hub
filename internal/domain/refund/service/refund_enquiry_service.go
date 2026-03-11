package service

import (
	"context"
	"time"

	"lending-hub-service/internal/domain/refund/entity"
	"lending-hub-service/internal/domain/refund/port"
	"lending-hub-service/internal/infrastructure/observability/metrics"
	baseLogger "lending-hub-service/pkg/logger"
)

// RefundEnquiryService handles enquiry-based resolution of refund states
type RefundEnquiryService struct {
	gateway    port.RefundGateway
	repository port.RefundRepository
	cache      port.RefundCache
	mc         metrics.MetricsClient
	logger     *baseLogger.Logger
	enquirySLA time.Duration
}

// NewRefundEnquiryService creates a new RefundEnquiryService
func NewRefundEnquiryService(
	gateway port.RefundGateway,
	repository port.RefundRepository,
	cache port.RefundCache,
	mc metrics.MetricsClient,
	logger *baseLogger.Logger,
	enquirySLA time.Duration,
) *RefundEnquiryService {
	return &RefundEnquiryService{
		gateway:    gateway,
		repository: repository,
		cache:      cache,
		mc:         mc,
		logger:     logger,
		enquirySLA: enquirySLA,
	}
}

// ResolveRefundState calls enquiry using order's loanId and updates refund
func (s *RefundEnquiryService) ResolveRefundState(ctx context.Context, refund *entity.Refund) error {
	merchantTxnID := refund.LoanID
	if merchantTxnID == "" && refund.ProviderMerchantTxnID != nil {
		merchantTxnID = *refund.ProviderMerchantTxnID
	}
	if merchantTxnID == "" {
		s.logger.Warn("cannot resolve refund: missing loanId",
			baseLogger.RefundID(refund.RefundID),
		)
		return nil
	}

	resp, err := s.gateway.EnquireRefund(ctx, merchantTxnID)
	if err != nil {
		s.logger.Error("enquiry API failed",
			baseLogger.RefundID(refund.RefundID),
			baseLogger.ErrorCode(err.Error()),
		)
		return nil // soft fail
	}

	var refundTxn *port.EnquiryTransaction
	for i := range resp.Transactions {
		txn := &resp.Transactions[i]
		if txn.TxnType != "REFUND" {
			continue
		}
		if txn.TxnRefNo != "" && txn.TxnRefNo == refund.RefundID {
			refundTxn = txn
			break
		}
		if txn.TxnRefNo == "" && refund.ProviderRefundTxnID != nil && txn.LpTxnID == *refund.ProviderRefundTxnID {
			refundTxn = txn
			break
		}
	}

	refund.RecordEnquiry()

	if refundTxn == nil {
		if time.Since(refund.CreatedAt) > s.enquirySLA {
			refund.MarkFailed("REFUND_NOT_FOUND_IN_ENQUIRY", "Refund not visible at lender after SLA")
		} else {
			if refund.Status != entity.RefundStatusUnknown {
				refund.MarkUnknown("Refund not yet visible in enquiry")
			}
		}
	} else {
		switch refundTxn.Status {
		case "SUCCESS":
			refund.MarkSuccess(refundTxn.LpTxnID, "", refundTxn.RespMessage)
			s.mc.Count(metrics.MetricRefundCompleted, 1, []string{"provider:" + refund.Lender})
		case "FAILED":
			refund.MarkFailed(refundTxn.Status, refundTxn.RespMessage)
		case "PROCESSING":
			refund.MarkProcessing(refundTxn.RespMessage)
		default:
			refund.MarkFailed(refundTxn.Status, refundTxn.RespMessage)
		}
	}

	if err := s.repository.Update(ctx, refund); err != nil {
		return err
	}

	ttl := 30 * time.Second
	if refund.Status.IsTerminal() {
		ttl = 5 * time.Minute
	}
	if err := s.cache.Set(ctx, refund.RefundID, refund.Status, ttl); err != nil {
		s.logger.Warn("cache.Set failed after ResolveRefundState", baseLogger.RefundID(refund.RefundID), baseLogger.Module("refund"))
	}

	s.logger.Info("refund state resolved via enquiry",
		baseLogger.RefundID(refund.RefundID),
		baseLogger.Status(string(refund.Status)),
	)

	return nil
}
