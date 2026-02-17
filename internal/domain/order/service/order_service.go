package service

import (
	"context"
	"encoding/json"
	"time"

	req "lending-hub-service/internal/domain/order/dto/request"
	res "lending-hub-service/internal/domain/order/dto/response"
	"lending-hub-service/internal/domain/order/entity"
	"lending-hub-service/internal/domain/order/port"
	profileService "lending-hub-service/internal/domain/profile/service"
	"lending-hub-service/internal/shared/idgen"
	sharedErrors "lending-hub-service/internal/shared/errors"
)

// OrderService handles order operations
type OrderService struct {
	orderRepo      port.OrderRepository
	mappingRepo    port.PaymentMappingRepository
	idempotency    *IdempotencyService
	gateway        port.OrderGateway
	profileUpdater *profileService.ProfileUpdater
	publisher      port.OrderEventPublisher
	idgen          *idgen.IDGenerator
}

// NewOrderService creates a new OrderService
func NewOrderService(
	orderRepo port.OrderRepository,
	mappingRepo port.PaymentMappingRepository,
	idempotency *IdempotencyService,
	gateway port.OrderGateway,
	profileUpdater *profileService.ProfileUpdater,
	publisher port.OrderEventPublisher,
	idgen *idgen.IDGenerator,
) *OrderService {
	return &OrderService{
		orderRepo:      orderRepo,
		mappingRepo:    mappingRepo,
		idempotency:    idempotency,
		gateway:        gateway,
		profileUpdater: profileUpdater,
		publisher:      publisher,
		idgen:          idgen,
	}
}

// CreateOrder creates a new order with idempotency
func (s *OrderService) CreateOrder(ctx context.Context, req req.CreateOrderRequest) (*res.OrderResponse, error) {
	// Generate payment ID if not provided
	paymentID := req.PaymentID
	if paymentID == "" {
		paymentID = s.idgen.PaymentID()
	}

	// Compute request hash
	requestHash := s.idempotency.ComputeHash(req)

	// Acquire idempotency key
	result, key, err := s.idempotency.Acquire(ctx, paymentID, requestHash)
	if err != nil {
		return nil, err
	}

	// Handle idempotency results
	switch result {
	case IdempotencyDuplicate:
		// Return cached response
		var cachedResponse res.OrderResponse
		if err := json.Unmarshal(key.ResponsePayload, &cachedResponse); err != nil {
			return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to unmarshal cached response")
		}
		return &cachedResponse, nil

	case IdempotencyConflict:
		return nil, sharedErrors.New(sharedErrors.CodeIdempotencyConflict, 409, "order creation already in progress")

	case IdempotencyMismatch:
		return nil, sharedErrors.New(sharedErrors.CodeHashMismatch, 422, "request hash mismatch: same paymentId with different request body")

	case IdempotencyNew:
		// Proceed with order creation
		break
	}

	// Call gateway to create order
	gatewayResp, err := s.gateway.CreateOrder(ctx, req)
	if err != nil {
		// Mark idempotency as failed
		s.idempotency.Fail(ctx, paymentID)
		return nil, err
	}

	// Serialize EMI selection
	emiPlanBytes, _ := json.Marshal(req.EmiSelection)

	// Create order entity
	order := &entity.Order{
		PaymentID:            paymentID,
		UserID:               req.UserID,
		MerchantID:           req.MerchantID,
		Lender:               "LAZYPAY", // TODO: make configurable
		Amount:               req.Amount,
		Currency:             req.Currency,
		Status:               entity.OrderStatus(gatewayResp.Status),
		ReturnURL:            &req.ReturnURL,
		EMIPlan:              emiPlanBytes,
		LenderOrderID:        gatewayResp.LenderOrderID,
		LenderMerchantTxnID:  gatewayResp.LenderOrderID, // Use lenderOrderID as merchant txn ID
		LastErrorCode:        gatewayResp.ErrorCode,
		LastErrorMessage:     gatewayResp.ErrorMessage,
		CreatedAt:            time.Now().UTC(),
		UpdatedAt:            time.Now().UTC(),
	}

	// Persist order
	if err := s.orderRepo.Create(ctx, order); err != nil {
		s.idempotency.Fail(ctx, paymentID)
		return nil, err
	}

	// Create payment mapping
	if gatewayResp.LenderOrderID != nil {
		mapping := &entity.PaymentMapping{
			PaymentID:            paymentID,
			UserID:               req.UserID,
			Lender:               order.Lender,
			LenderMerchantTxnID:  *gatewayResp.LenderOrderID,
			LenderOrderID:        gatewayResp.LenderOrderID,
			CreatedAt:            time.Now().UTC(),
			UpdatedAt:            time.Now().UTC(),
		}
		if err := s.mappingRepo.Create(ctx, mapping); err != nil {
			// Log error but don't fail the request
		}
	}

	// Build response
	response := &res.OrderResponse{
		PaymentID:     paymentID,
		Status:        gatewayResp.Status,
		LenderOrderID: gatewayResp.LenderOrderID,
		RedirectURL:   gatewayResp.RedirectURL,
		ErrorCode:     gatewayResp.ErrorCode,
		ErrorMessage:  gatewayResp.ErrorMessage,
	}

	// Marshal response for idempotency cache
	responseBytes, _ := json.Marshal(response)

	// Mark idempotency as completed
	if err := s.idempotency.Complete(ctx, paymentID, responseBytes, gatewayResp.LenderOrderID); err != nil {
		// Log error but don't fail the request
	}

	// Publish event
	s.publisher.Publish(ctx, &port.OrderEvent{
		Type:          port.EventOrderCreated,
		PaymentID:     paymentID,
		UserID:        req.UserID,
		MerchantID:    req.MerchantID,
		Lender:        order.Lender,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Status:        gatewayResp.Status,
		LenderOrderID: gatewayResp.LenderOrderID,
	})

	return response, nil
}

// GetOrderStatus retrieves order status
func (s *OrderService) GetOrderStatus(ctx context.Context, paymentID string) (*res.OrderStatusResponse, error) {
	order, err := s.orderRepo.GetByPaymentID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, sharedErrors.New(sharedErrors.CodeOrderNotFound, 404, "order not found")
	}

	response := &res.OrderStatusResponse{
		PaymentID:            order.PaymentID,
		UserID:               order.UserID,
		MerchantID:           order.MerchantID,
		Amount:               order.Amount,
		Currency:             order.Currency,
		Status:               string(order.Status),
		LenderOrderID:        order.LenderOrderID,
		LenderMerchantTxnID:  order.LenderMerchantTxnID,
		LenderLastStatus:     order.LenderLastStatus,
		LenderLastTxnID:      order.LenderLastTxnID,
		LenderLastTxnStatus:  order.LenderLastTxnStatus,
		LenderLastTxnMessage: order.LenderLastTxnMessage,
		LastErrorCode:        order.LastErrorCode,
		LastErrorMessage:     order.LastErrorMessage,
		CreatedAt:            order.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            order.UpdatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// ProcessCallback processes an order callback event
func (s *OrderService) ProcessCallback(ctx context.Context, req req.OrderCallbackRequest) error {
	// Parse event time
	eventTime, err := time.Parse(time.RFC3339, req.EventTime)
	if err != nil {
		return sharedErrors.New(sharedErrors.CodeInvalidRequest, 400, "invalid eventTime format")
	}

	// Get order for update (FOR UPDATE lock)
	order, err := s.orderRepo.GetForUpdate(ctx, req.PaymentID)
	if err != nil {
		return err
	}
	if order == nil {
		return sharedErrors.New(sharedErrors.CodeOrderNotFound, 404, "order not found")
	}

	// If already terminal, ignore (idempotent)
	if order.Status.IsTerminal() {
		return nil
	}

	// Update order status
	newStatus := entity.OrderStatus(req.Status)
	order.Status = newStatus
	order.LenderLastStatus = &req.Status
	order.LenderLastTxnID = req.LenderTxnID
	order.LenderLastTxnStatus = &req.Status
	order.LenderLastTxnMessage = req.ErrorMessage
	order.LenderLastTxnTime = &eventTime
	order.LastErrorCode = req.ErrorCode
	order.LastErrorMessage = req.ErrorMessage
	order.UpdatedAt = time.Now().UTC()

	// Update lender order ID if provided
	if req.LenderOrderID != nil {
		order.LenderOrderID = req.LenderOrderID
	}

	// If SUCCESS, update profile limit (deduct available limit)
	if newStatus == entity.OrderSuccess {
		// Deduct amount from available limit
		// Get current profile to check available limit
		// For now, we'll just update the limit (in production, you'd check first)
		newAvailable := 0.0 // TODO: calculate from current available - amount
		if err := s.profileUpdater.UpdateLimit(ctx, order.UserID, order.Lender, newAvailable); err != nil {
			// Log error but don't fail callback
		}
	}

	// Update order
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return err
	}

	// Publish event
	eventType := port.EventOrderCompleted
	if newStatus == entity.OrderFailed {
		eventType = port.EventOrderFailed
	}

	s.publisher.Publish(ctx, &port.OrderEvent{
		Type:          eventType,
		PaymentID:     order.PaymentID,
		UserID:        order.UserID,
		MerchantID:    order.MerchantID,
		Lender:        order.Lender,
		Amount:        order.Amount,
		Currency:      order.Currency,
		Status:        req.Status,
		LenderOrderID: order.LenderOrderID,
		LenderTxnID:   req.LenderTxnID,
		ErrorCode:     req.ErrorCode,
		ErrorMessage:  req.ErrorMessage,
	})

	return nil
}
