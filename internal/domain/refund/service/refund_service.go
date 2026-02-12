package service

import (
	"context"
	"time"

	req "lending-hub-service/internal/domain/refund/dto/request"
	res "lending-hub-service/internal/domain/refund/dto/response"
	"lending-hub-service/internal/domain/refund/entity"
	"lending-hub-service/internal/domain/refund/port"
	orderEntity "lending-hub-service/internal/domain/order/entity"
	orderPort "lending-hub-service/internal/domain/order/port"
	profileService "lending-hub-service/internal/domain/profile/service"
	sharedErrors "lending-hub-service/internal/shared/errors"
)

// RefundService handles refund operations
type RefundService struct {
	refundRepo     port.RefundRepository
	orderRepo      orderPort.OrderRepository
	gateway        port.RefundGateway
	profileUpdater *profileService.ProfileUpdater
}

// NewRefundService creates a new RefundService
func NewRefundService(
	refundRepo port.RefundRepository,
	orderRepo orderPort.OrderRepository,
	gateway port.RefundGateway,
	profileUpdater *profileService.ProfileUpdater,
) *RefundService {
	return &RefundService{
		refundRepo:     refundRepo,
		orderRepo:      orderRepo,
		gateway:        gateway,
		profileUpdater: profileUpdater,
	}
}

// CreateRefund creates a new refund with validation
func (s *RefundService) CreateRefund(ctx context.Context, req req.CreateRefundRequest) (*res.RefundResponse, error) {
	// Fetch order with FOR UPDATE lock
	order, err := s.orderRepo.GetForUpdate(ctx, req.PaymentID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, sharedErrors.New(sharedErrors.CodeOrderNotFound, 404, "order not found")
	}

	// Validate order status is SUCCESS
	if order.Status != orderEntity.OrderSuccess {
		return nil, sharedErrors.New(sharedErrors.CodeInvalidRequest, 400, "can only refund orders with SUCCESS status")
	}

	// Fetch existing refunds for this payment
	existingRefunds, err := s.refundRepo.ListByPaymentID(ctx, req.PaymentID)
	if err != nil {
		return nil, err
	}

	// Calculate total refunded amount (only SUCCESS refunds)
	totalRefunded := 0.0
	for _, refund := range existingRefunds {
		if refund.Status == entity.RefundSuccess {
			totalRefunded += refund.Amount
		}
	}

	// Validate: new refund amount + existing refunds <= order amount
	if totalRefunded+req.Amount > order.Amount {
		return nil, sharedErrors.New(sharedErrors.CodeRefundExceedsOrder, 422, "refund amount exceeds order amount")
	}

	// Call gateway to process refund
	gatewayResp, err := s.gateway.ProcessRefund(ctx, req)
	if err != nil {
		return nil, err
	}

	// Parse refund reason
	var reason *entity.RefundReason
	if req.Reason != "" {
		reasonVal := entity.RefundReason(req.Reason)
		reason = &reasonVal
	}

	// Create refund entity
	refund := &entity.Refund{
		RefundID:      req.RefundID,
		PaymentID:     req.PaymentID,
		UserID:        order.UserID,
		Lender:        order.Lender,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Status:        entity.RefundStatus(gatewayResp.Status),
		Reason:        reason,
		LenderRefID:   gatewayResp.LenderRefID,
		LenderStatus:  &gatewayResp.Status,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	// Persist refund
	if err := s.refundRepo.Create(ctx, refund); err != nil {
		return nil, err
	}

	return &res.RefundResponse{
		RefundID:  refund.RefundID,
		PaymentID: refund.PaymentID,
		Provider:  gatewayResp.Provider,
		Status:    gatewayResp.Status,
		Amount:    refund.Amount,
		Currency:  refund.Currency,
	}, nil
}

// ProcessCallback processes a refund callback event
func (s *RefundService) ProcessCallback(ctx context.Context, req req.RefundCallbackRequest) error {
	// Parse event time (validate format)
	_, err := time.Parse(time.RFC3339, req.EventTime)
	if err != nil {
		return sharedErrors.New(sharedErrors.CodeInvalidRequest, 400, "invalid eventTime format")
	}

	// Get refund
	refund, err := s.refundRepo.GetByRefundID(ctx, req.RefundID)
	if err != nil {
		return err
	}
	if refund == nil {
		return sharedErrors.New(sharedErrors.CodeRefundNotFound, 404, "refund not found")
	}

	// Validate payment ID matches
	if refund.PaymentID != req.PaymentID {
		return sharedErrors.New(sharedErrors.CodeInvalidRequest, 400, "paymentId mismatch")
	}

	// Update refund status
	newStatus := entity.RefundStatus(req.Status)
	refund.Status = newStatus
	refund.LenderRefID = req.LenderRefID
	refund.LenderStatus = &req.Status
	refund.LenderMessage = req.LenderMessage
	refund.UpdatedAt = time.Now().UTC()

	// If SUCCESS, restore available limit
	if newStatus == entity.RefundSuccess {
		// Add refund amount back to available limit
		if err := s.profileUpdater.AddToLimit(ctx, refund.UserID, refund.Lender, refund.Amount); err != nil {
			// Log error but don't fail callback
		}

		// Check if all refunds for this order are SUCCESS and total equals order amount
		allRefunds, err := s.refundRepo.ListByPaymentID(ctx, req.PaymentID)
		if err == nil {
			totalRefunded := 0.0
			allSuccess := true
			for _, r := range allRefunds {
				if r.Status == entity.RefundSuccess {
					totalRefunded += r.Amount
				} else if r.Status != entity.RefundSuccess {
					allSuccess = false
				}
			}

			// Get order to check amount
			order, err := s.orderRepo.GetForUpdate(ctx, req.PaymentID)
			if err == nil && order != nil {
				if allSuccess && totalRefunded >= order.Amount {
					// Mark order as fully refunded
					order.Status = orderEntity.OrderRefunded
					order.UpdatedAt = time.Now().UTC()
					s.orderRepo.Update(ctx, order)
				}
			}
		}
	}

	// Update refund
	if err := s.refundRepo.Update(ctx, refund); err != nil {
		return err
	}

	return nil
}
