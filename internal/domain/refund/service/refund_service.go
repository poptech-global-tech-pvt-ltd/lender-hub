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
	sharedErrors "lending-hub-service/internal/shared/errors"
)

// RefundService handles refund operations
type RefundService struct {
	refundRepo     port.RefundRepository
	orderRepo      orderPort.OrderRepository
	mappingRepo    orderPort.PaymentMappingRepository
	gateway        port.RefundGateway
	profileUpdater *profileService.ProfileUpdater
}

// NewRefundService creates a new RefundService
func NewRefundService(
	refundRepo port.RefundRepository,
	orderRepo orderPort.OrderRepository,
	mappingRepo orderPort.PaymentMappingRepository,
	gateway port.RefundGateway,
	profileUpdater *profileService.ProfileUpdater,
) *RefundService {
	return &RefundService{
		refundRepo:     refundRepo,
		orderRepo:      orderRepo,
		mappingRepo:    mappingRepo,
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
		if refund.Status == entity.RefundStatusSuccess {
			totalRefunded += refund.Amount
		}
	}

	// Validate: new refund amount + existing refunds <= order amount
	if totalRefunded+req.Amount > order.Amount {
		return nil, sharedErrors.New(sharedErrors.CodeRefundExceedsOrder, 422, "refund amount exceeds order amount")
	}

	// Check for duplicate refund using providerRefundRefID (idempotency key)
	// providerRefundRefID = req.RefundID (used as idempotency key)
	existingRefund, err := s.refundRepo.GetByProviderRefID(ctx, order.Lender, req.RefundID)
	if err != nil {
		return nil, err
	}
	if existingRefund != nil {
		// Return existing refund
		return &res.RefundResponse{
			RefundID:  existingRefund.RefundID,
			PaymentID: existingRefund.PaymentID,
			Provider:  "LAZYPAY",
			Status:    string(existingRefund.Status),
			Amount:    existingRefund.Amount,
			Currency:  existingRefund.Currency,
		}, nil
	}

	// Look up merchantTxnId from paymentId mapping
	mapping, err := s.mappingRepo.GetByPaymentID(ctx, req.PaymentID)
	if err != nil || mapping == nil {
		return nil, sharedErrors.New(sharedErrors.CodeOrderNotFound, 404, "order mapping not found")
	}
	merchantTxnID := mapping.LenderMerchantTxnID

	// Call gateway to process refund (returns response + generated refundTxnId)
	gatewayResp, refundTxnID, err := s.gateway.ProcessRefund(ctx, merchantTxnID, req.Amount, req.Currency)
	if err != nil {
		// Check if error is LP_DUPLICATE_REFUND
		if domainErr, ok := err.(*sharedErrors.DomainError); ok && domainErr.Code == sharedErrors.CodeDuplicateRefund {
			// Duplicate refund detected - call enquiry API to get existing refund status
			enquiryResp, enquiryErr := s.gateway.EnquireRefund(ctx, req.PaymentID)
			if enquiryErr != nil {
				// If enquiry fails, return the duplicate error
				return nil, err
			}

			// Find REFUND transaction matching our providerRefundRefID
			var refundTxn *port.EnquiryTransaction
			for i := range enquiryResp.Transactions {
				txn := &enquiryResp.Transactions[i]
				if txn.TxnType == "REFUND" && txn.TxnRefNo == req.RefundID {
					refundTxn = txn
					break
				}
			}

			if refundTxn != nil {
				// Found existing refund via enquiry - create refund entity from enquiry data
				var reason *entity.RefundReason
				if req.Reason != "" {
					reasonVal := entity.RefundReason(req.Reason)
					reason = &reasonVal
				}

				refundStatus := entity.RefundStatusPending
				if refundTxn.Status == "SUCCESS" {
					refundStatus = entity.RefundStatusSuccess
				} else if refundTxn.Status == "FAILED" {
					refundStatus = entity.RefundStatusFailed
				} else if refundTxn.Status == "PROCESSING" {
					refundStatus = entity.RefundStatusProcessing
				}

				refund := &entity.Refund{
					RefundID:              req.RefundID,
					PaymentID:             req.PaymentID,
					UserID:                order.UserID,
					Lender:                order.Lender,
					Amount:                req.Amount,
					Currency:              req.Currency,
					Status:                refundStatus,
					Reason:                reason,
					ProviderRefundRefID:   req.RefundID,
					ProviderMerchantTxnID: &req.PaymentID,
					ProviderRefundTxnID:   &refundTxn.LpTxnID,
					LenderStatus:          &refundTxn.Status,
					LenderMessage:         &refundTxn.RespMessage,
					CreatedAt:             time.Now().UTC(),
					UpdatedAt:             time.Now().UTC(),
				}

				// Persist refund
				if err := s.refundRepo.Create(ctx, refund); err != nil {
					return nil, err
				}

				return &res.RefundResponse{
					RefundID:  refund.RefundID,
					PaymentID: refund.PaymentID,
					Provider:  "LAZYPAY",
					Status:    string(refund.Status),
					Amount:    refund.Amount,
					Currency:  refund.Currency,
				}, nil
			}
		}
		// Return original error if not duplicate or enquiry failed
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
		RefundID:              req.RefundID,     // Caller's reference
		PaymentID:             req.PaymentID,
		UserID:                order.UserID,
		Lender:                order.Lender,
		Amount:                req.Amount,
		Currency:              req.Currency,
		Status:                entity.RefundStatus(gatewayResp.Status),
		Reason:                reason,
		ProviderRefundRefID:   req.RefundID, // idempotency key
		ProviderMerchantTxnID: &merchantTxnID,
		ProviderRefundTxnID:   &refundTxnID, // Our generated refundTxnId sent to Lazypay
		LenderRefID:           gatewayResp.LenderRefID,
		LenderStatus:          &gatewayResp.Status,
		CreatedAt:             time.Now().UTC(),
		UpdatedAt:             time.Now().UTC(),
	}

	// Persist refund
	if err := s.refundRepo.Create(ctx, refund); err != nil {
		return nil, err
	}

	return &res.RefundResponse{
		RefundID:  refund.RefundID, // Caller's reference
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
	if newStatus == entity.RefundStatusSuccess {
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
				if r.Status == entity.RefundStatusSuccess {
					totalRefunded += r.Amount
				} else if r.Status != entity.RefundStatusSuccess {
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

// GetRefundStatus retrieves refund status by refund ID
func (s *RefundService) GetRefundStatus(ctx context.Context, refundID string) (*res.RefundResponse, error) {
	refund, err := s.refundRepo.GetByRefundID(ctx, refundID)
	if err != nil {
		return nil, err
	}
	if refund == nil {
		return nil, sharedErrors.New(sharedErrors.CodeRefundNotFound, 404, "refund not found")
	}

	status := string(refund.Status)
	return &res.RefundResponse{
		RefundID:    refund.RefundID,
		PaymentID:   refund.PaymentID,
		Provider:    "LAZYPAY",
		Status:      status,
		Amount:      refund.Amount,
		Currency:    refund.Currency,
		LenderRefID: refund.LenderRefID,
	}, nil
}
