package service

import (
	"context"
	"time"

	"lending-hub-service/internal/domain/refund/entity"
	"lending-hub-service/internal/domain/refund/port"
	baseLogger "lending-hub-service/pkg/logger"
	"lending-hub-service/internal/infrastructure/observability/metrics"
)

// RefundEnquiryService handles enquiry-based resolution of refund states
type RefundEnquiryService struct {
	gateway    port.RefundGateway
	repository port.RefundRepository
	mc         metrics.MetricsClient
	logger     *baseLogger.Logger
	enquirySLA time.Duration
}

// NewRefundEnquiryService creates a new RefundEnquiryService
func NewRefundEnquiryService(
	gateway port.RefundGateway,
	repository port.RefundRepository,
	mc metrics.MetricsClient,
	logger *baseLogger.Logger,
	enquirySLA time.Duration,
) *RefundEnquiryService {
	return &RefundEnquiryService{
		gateway:    gateway,
		repository: repository,
		mc:         mc,
		logger:     logger,
		enquirySLA: enquirySLA,
	}
}

// ResolveRefundState calls enquiry API and updates refund based on response
func (s *RefundEnquiryService) ResolveRefundState(ctx context.Context, refund *entity.Refund) error {
	s.logger.Info("resolving refund via enquiry",
		baseLogger.RefundID(refund.RefundID),
		baseLogger.Status(string(refund.Status)),
	)

	if refund.ProviderMerchantTxnID == nil || *refund.ProviderMerchantTxnID == "" {
		s.logger.Warn("cannot resolve refund: missing provider_merchant_txn_id",
			baseLogger.RefundID(refund.RefundID),
		)
		return nil
	}

	// Call Lazypay enquiry API
	resp, err := s.gateway.EnquireRefund(ctx, *refund.ProviderMerchantTxnID)
	if err != nil {
		s.logger.Error("enquiry API failed",
			baseLogger.RefundID(refund.RefundID),
			baseLogger.ErrorCode(err.Error()),
		)
		return err
	}

	// Find REFUND transaction matching our provider_refund_ref_id
	var refundTxn *port.EnquiryTransaction
	for i := range resp.Transactions {
		txn := &resp.Transactions[i]
		if txn.TxnType == "REFUND" && txn.TxnRefNo == refund.ProviderRefundRefID {
			refundTxn = txn
			break
		}
	}

	refund.RecordEnquiry()

	if refundTxn == nil {
		// Refund not found in enquiry
		if time.Since(refund.CreatedAt) > s.enquirySLA {
			// Past SLA, mark as FAILED
			refund.MarkFailed("REFUND_NOT_FOUND_IN_ENQUIRY",
				"Refund not visible at lender after SLA")
			s.logger.Warn("refund not found after SLA, marking FAILED",
				baseLogger.RefundID(refund.RefundID),
			)
		} else {
			// Still within SLA, keep current state (or mark UNKNOWN if PENDING)
			if refund.Status == entity.RefundStatusPending {
				refund.MarkUnknown("Refund not yet visible in enquiry")
			}
			s.logger.Info("refund not found in enquiry, within SLA",
				baseLogger.RefundID(refund.RefundID),
			)
		}
	} else {
		// Refund found in enquiry
		switch refundTxn.Status {
		case "SUCCESS":
			refund.MarkSuccess(refundTxn.LpTxnID, "", refundTxn.RespMessage)
			s.mc.Count(metrics.MetricRefundCompleted, 1, []string{
				"provider:" + refund.Lender,
			})
			s.logger.Info("refund resolved as SUCCESS via enquiry",
				baseLogger.RefundID(refund.RefundID),
			)
		case "FAILED":
			refund.MarkFailed(refundTxn.Status, refundTxn.RespMessage)
			s.logger.Info("refund resolved as FAILED via enquiry",
				baseLogger.RefundID(refund.RefundID),
			)
		case "PROCESSING":
			refund.MarkProcessing(refundTxn.RespMessage)
			s.logger.Info("refund still PROCESSING per enquiry",
				baseLogger.RefundID(refund.RefundID),
			)
		default:
			refund.MarkFailed(refundTxn.Status, refundTxn.RespMessage)
			s.logger.Warn("unknown enquiry status, marking FAILED",
				baseLogger.RefundID(refund.RefundID),
				baseLogger.Status(refundTxn.Status),
			)
		}
	}

	// Persist updated state
	if err := s.repository.Update(ctx, refund); err != nil {
		s.logger.Error("failed to update refund after enquiry",
			baseLogger.RefundID(refund.RefundID),
			baseLogger.ErrorCode(err.Error()),
		)
		return err
	}

	// TODO: If SUCCESS, trigger async limit restoration (Kafka event or direct call)

	return nil
}
