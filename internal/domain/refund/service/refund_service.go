package service

import (
	"context"
	"time"

	orderEntity "lending-hub-service/internal/domain/order/entity"
	orderPort "lending-hub-service/internal/domain/order/port"
	profileService "lending-hub-service/internal/domain/profile/service"
	req "lending-hub-service/internal/domain/refund/dto/request"
	res "lending-hub-service/internal/domain/refund/dto/response"
	"lending-hub-service/internal/domain/refund/entity"
	"lending-hub-service/internal/domain/refund/port"
	"lending-hub-service/internal/infrastructure/observability/metrics"
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/pkg/idgen"
	baseLogger "lending-hub-service/pkg/logger"
)

const lenderLAZYPAY = "LAZYPAY"

// RefundService handles refund operations
type RefundService struct {
	repo           port.RefundRepository
	orderRepo      orderPort.OrderRepository
	orderGateway   orderPort.OrderGateway
	gateway        port.RefundGateway
	cache          port.RefundCache
	enquiryService *RefundEnquiryService
	profileUpdater *profileService.ProfileUpdater
	mc             metrics.MetricsClient
	logger         *baseLogger.Logger
	idgen          *idgen.Generator
}

// NewRefundService creates a new RefundService
func NewRefundService(
	repo port.RefundRepository,
	orderRepo orderPort.OrderRepository,
	orderGateway orderPort.OrderGateway,
	gateway port.RefundGateway,
	cache port.RefundCache,
	enquiryService *RefundEnquiryService,
	profileUpdater *profileService.ProfileUpdater,
	mc metrics.MetricsClient,
	logger *baseLogger.Logger,
	idgen *idgen.Generator,
) *RefundService {
	return &RefundService{
		repo:           repo,
		orderRepo:      orderRepo,
		orderGateway:   orderGateway,
		gateway:        gateway,
		cache:          cache,
		enquiryService: enquiryService,
		profileUpdater: profileUpdater,
		mc:             mc,
		logger:         logger,
		idgen:          idgen,
	}
}

// CreateRefund creates a new refund
func (s *RefundService) CreateRefund(ctx context.Context, r req.CreateRefundRequest) (*res.RefundResponse, error) {
	order, err := s.orderRepo.GetByPaymentID(ctx, r.PaymentID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, sharedErrors.New(sharedErrors.CodeOrderNotFound, 404, "order not found")
	}
	// Refresh order from Lazypay if non-terminal (e.g. PENDING in DB but SUCCESS at Lazypay)
	if err := s.resolveOrderFromEnquiry(ctx, order); err == nil {
		if reloaded, _ := s.orderRepo.GetByPaymentID(ctx, r.PaymentID); reloaded != nil {
			order = reloaded
		}
	}
	if !order.Status.IsRefundable() {
		return nil, sharedErrors.New(sharedErrors.CodeOrderNotRefundable, 400, "can only refund orders with SUCCESS or COMPLETE status")
	}

	loanID := ""
	if order.LenderMerchantTxnID != nil {
		loanID = *order.LenderMerchantTxnID
	}
	if loanID == "" {
		return nil, sharedErrors.New(sharedErrors.CodeOrderNotFound, 404, "order loanId not found")
	}

	existing, _ := s.repo.GetByPaymentRefundID(ctx, lenderLAZYPAY, r.PaymentRefundID)
	if existing != nil {
		if existing.Status.IsTerminal() {
			return MapRefundToResponse(existing), nil
		}
		_ = s.enquiryService.ResolveRefundState(ctx, existing)
		reloaded, _ := s.repo.GetByPaymentRefundID(ctx, lenderLAZYPAY, r.PaymentRefundID)
		if reloaded != nil {
			return MapRefundToResponse(reloaded), nil
		}
		return MapRefundToResponse(existing), nil
	}

	refundID := s.idgen.RefundID()
	refund := &entity.Refund{
		RefundID:              refundID,
		PaymentRefundID:       r.PaymentRefundID,
		PaymentID:             r.PaymentID,
		LoanID:                loanID,
		UserID:                order.UserID,
		Lender:                order.Lender,
		Amount:                r.Amount,
		Currency:              r.Currency,
		Status:                entity.RefundStatusPending,
		ProviderMerchantTxnID: &loanID,
		ProviderRefundRefID:   refundID,
		CreatedAt:             time.Now().UTC(),
		UpdatedAt:             time.Now().UTC(),
	}
	if r.Reason != "" {
		reason := entity.RefundReason(r.Reason)
		refund.Reason = &reason
	}

	if err := s.repo.Create(ctx, refund); err != nil {
		return nil, err
	}

	s.mc.Count(metrics.MetricRefundInitiated, 1, []string{"provider:" + order.Lender})

	gatewayReq := port.ProcessRefundRequest{
		MerchantTxnID: loanID,
		Amount:        r.Amount,
		Currency:      r.Currency,
		RefundTxnID:   refundID,
	}
	resp, err := s.gateway.ProcessRefund(ctx, gatewayReq)

	if err != nil {
		lenderStatus, lenderMsg := "API_ERROR", "Refund API call failed"
		if de, ok := err.(*sharedErrors.DomainError); ok {
			lenderStatus, lenderMsg = de.Code, de.Message
		}
		refund.MarkFailed(lenderStatus, lenderMsg)
		_ = s.repo.Update(ctx, refund)
		return nil, err
	}

	if resp.IsTimeout {
		refund.MarkUnknown("Refund initiation timed out")
		_ = s.repo.Update(ctx, refund)
		_ = s.cache.Set(ctx, refundID, entity.RefundStatusUnknown, 30*time.Second)
		return MapRefundToResponse(refund), nil
	}

	if resp.ErrorCode == "LPDUPLICATEREFUND" {
		s.logger.Info("LPDUPLICATEREFUND received, triggering enquiry",
			baseLogger.RefundID(refundID))
		_ = s.enquiryService.ResolveRefundState(ctx, refund)
		reloaded, _ := s.repo.GetByRefundID(ctx, refundID)
		if reloaded != nil {
			return MapRefundToResponse(reloaded), nil
		}
		return MapRefundToResponse(refund), nil
	}

	if resp.Status == "REFUND_SUCCESS" {
		refund.MarkSuccess(resp.LpTxnID, resp.ParentTxnID, resp.RespMessage)
		s.mc.Count(metrics.MetricRefundCompleted, 1, []string{"provider:" + order.Lender, "trigger:direct"})
		go s.profileUpdater.AddToLimit(context.Background(), order.UserID, order.Lender, r.Amount)
	} else {
		refund.MarkFailed(resp.Status, resp.RespMessage)
	}

	_ = s.repo.Update(ctx, refund)
	ttl := 30 * time.Second
	if refund.Status.IsTerminal() {
		ttl = 5 * time.Minute
	}
	_ = s.cache.Set(ctx, refundID, refund.Status, ttl)

	s.logger.Info("refund created",
		baseLogger.RefundID(refundID),
		baseLogger.Status(string(refund.Status)))

	return MapRefundToResponse(refund), nil
}

// resolveOrderFromEnquiry calls Lazypay order enquiry and updates order if non-terminal
func (s *RefundService) resolveOrderFromEnquiry(ctx context.Context, order *orderEntity.Order) error {
	if order.Status.IsTerminal() {
		return nil
	}
	loanID := ""
	if order.LenderMerchantTxnID != nil {
		loanID = *order.LenderMerchantTxnID
	}
	if loanID == "" {
		return nil
	}
	gatewayResp, err := s.orderGateway.GetOrderStatus(ctx, loanID)
	if err != nil {
		return err
	}
	// COMPLETE → SUCCESS for DB enum (lender_payment_status has no COMPLETE)
	newStatus := orderEntity.OrderStatus(gatewayResp.Status).OrDefault().NormalizeForDB()
	order.Status = newStatus
	if gatewayResp.LenderOrderID != "" {
		order.LenderOrderID = &gatewayResp.LenderOrderID
	}
	order.LenderLastStatus = &gatewayResp.Status
	order.UpdatedAt = time.Now().UTC()
	return s.orderRepo.Update(ctx, order)
}

// GetByPaymentRefundID retrieves by POP's paymentRefundId, triggers enquiry if non-terminal
func (s *RefundService) GetByPaymentRefundID(ctx context.Context, paymentRefundID string) (*entity.Refund, error) {
	refund, err := s.repo.GetByPaymentRefundID(ctx, lenderLAZYPAY, paymentRefundID)
	if err != nil || refund == nil {
		return nil, sharedErrors.New(sharedErrors.CodeRefundNotFound, 404, "refund not found")
	}
	refund.Status = refund.Status.OrDefault()
	if refund.Status.IsResolvable() {
		_ = s.enquiryService.ResolveRefundState(ctx, refund)
		refund, _ = s.repo.GetByPaymentRefundID(ctx, lenderLAZYPAY, paymentRefundID)
		if refund != nil {
			refund.Status = refund.Status.OrDefault()
		}
	}
	return refund, nil
}

// GetByRefundID retrieves by our refundId, triggers enquiry if non-terminal
func (s *RefundService) GetByRefundID(ctx context.Context, refundID string) (*entity.Refund, error) {
	refund, err := s.repo.GetByRefundID(ctx, refundID)
	if err != nil || refund == nil {
		return nil, sharedErrors.New(sharedErrors.CodeRefundNotFound, 404, "refund not found")
	}
	refund.Status = refund.Status.OrDefault()
	if refund.Status.IsResolvable() {
		_ = s.enquiryService.ResolveRefundState(ctx, refund)
		refund, _ = s.repo.GetByRefundID(ctx, refundID)
		if refund != nil {
			refund.Status = refund.Status.OrDefault()
		}
	}
	return refund, nil
}

// ListByPaymentID lists refunds for an order
func (s *RefundService) ListByPaymentID(ctx context.Context, paymentID string) ([]*entity.Refund, error) {
	refunds, err := s.repo.ListByPaymentID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	for _, r := range refunds {
		r.Status = r.Status.OrDefault()
	}
	return refunds, nil
}

// ListByUserID lists refunds for a user
func (s *RefundService) ListByUserID(ctx context.Context, userID string, page, perPage int) ([]*entity.Refund, int, error) {
	return s.repo.ListByUserID(ctx, userID, page, perPage)
}

// MapRefundToResponse maps entity to response DTO
func MapRefundToResponse(r *entity.Refund) *res.RefundResponse {
	resp := &res.RefundResponse{
		RefundID:        r.RefundID,
		PaymentRefundID: r.PaymentRefundID,
		PaymentID:       r.PaymentID,
		LoanID:          r.LoanID,
		Status:          string(r.Status.OrDefault()),
		Amount:          r.Amount,
		Currency:        r.Currency,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
	if r.ProviderRefundTxnID != nil {
		resp.ProviderRefundTxnID = *r.ProviderRefundTxnID
	}
	if r.Reason != nil {
		resp.Reason = string(*r.Reason)
	}
	if r.LenderStatus != nil {
		resp.LenderStatus = *r.LenderStatus
	}
	if r.LenderMessage != nil {
		resp.LenderMessage = *r.LenderMessage
	}
	resp.LastEnquiredAt = r.LastEnquiredAt
	return resp
}
