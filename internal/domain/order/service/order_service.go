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
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/pkg/idgen"
)

// OrderService handles order operations
type OrderService struct {
	orderRepo       port.OrderRepository
	mappingRepo     port.PaymentMappingRepository
	idempotency     *IdempotencyService
	gateway         port.OrderGateway
	profileUpdater  *profileService.ProfileUpdater
	publisher       port.OrderEventPublisher
	idgen           *idgen.Generator
	contactResolver *profileService.UserContactResolver
	merchantID      string // from config (subMerchantId), for DB NOT NULL
}

// NewOrderService creates a new OrderService
func NewOrderService(
	orderRepo port.OrderRepository,
	mappingRepo port.PaymentMappingRepository,
	idempotency *IdempotencyService,
	gateway port.OrderGateway,
	profileUpdater *profileService.ProfileUpdater,
	publisher port.OrderEventPublisher,
	idgen *idgen.Generator,
	contactResolver *profileService.UserContactResolver,
	merchantID string,
) *OrderService {
	if merchantID == "" {
		merchantID = "DEFAULT"
	}
	return &OrderService{
		orderRepo:       orderRepo,
		mappingRepo:     mappingRepo,
		idempotency:     idempotency,
		gateway:         gateway,
		profileUpdater:  profileUpdater,
		publisher:       publisher,
		idgen:           idgen,
		contactResolver: contactResolver,
		merchantID:      merchantID,
	}
}

// CreateOrder creates a new order with idempotency
func (s *OrderService) CreateOrder(ctx context.Context, req req.CreateOrderRequest) (*res.OrderResponse, error) {
	// Payment ID is ALWAYS server-generated — never from caller
	paymentID := s.idgen.PaymentID()

	// MerchantID and Source: use request or defaults
	merchantID := req.MerchantID
	if merchantID == "" {
		merchantID = s.merchantID
	}
	source := req.Source
	if source == "" {
		source = "CHECKOUT"
	}

	// EMI tenure: EMIPlan.Tenure or legacy EmiSelection.Tenure
	tenure := req.EMIPlan.Tenure
	if tenure == 0 && req.EmiSelection != nil {
		tenure = req.EmiSelection.Tenure
	}
	if tenure == 0 {
		return nil, sharedErrors.New(sharedErrors.CodeInvalidRequest, 400, "emiPlan.tenure or emiSelection.tenure required")
	}

	// Idempotency key = request hash (duplicate requests return cached response)
	requestHash := s.idempotency.ComputeHash(req)
	idempotencyKey := requestHash

	result, key, err := s.idempotency.Acquire(ctx, idempotencyKey, requestHash)
	if err != nil {
		return nil, err
	}

	switch result {
	case IdempotencyDuplicate:
		var cachedResponse res.OrderResponse
		if err := json.Unmarshal(key.ResponsePayload, &cachedResponse); err != nil {
			return nil, sharedErrors.New(sharedErrors.CodeInternalError, 500, "failed to unmarshal cached response")
		}
		return &cachedResponse, nil

	case IdempotencyConflict:
		return nil, sharedErrors.New(sharedErrors.CodeIdempotencyConflict, 409, "order creation already in progress")

	case IdempotencyMismatch:
		return nil, sharedErrors.New(sharedErrors.CodeHashMismatch, 422, "request hash mismatch: same idempotency key with different request body")

	case IdempotencyNew:
		break
	}

	// Resolve user contact (mobile + email) from userId
	contact, err := s.contactResolver.Resolve(ctx, req.UserID)
	if err != nil {
		s.idempotency.Fail(ctx, idempotencyKey)
		return nil, sharedErrors.New(sharedErrors.CodeUserContactNotFound, 422, "Unable to resolve user contact details (mobile required)")
	}

	merchantTxnID := s.idgen.MerchantTxnID()

	// Build EMI plans (single plan from tenure)
	emiType := "PAY_IN_PARTS"
	if req.EmiSelection != nil {
		emiType = req.EmiSelection.Type
	}
	emiPlans := []port.LPEmiPlan{
		{
			Tenure:                   tenure,
			Type:                     emiType,
			Emi:                      0,
			TotalPayableAmount:       0,
			InterestRate:             0,
			Principal:                0,
			TotalInterestAmount:      0,
			TotalProcessingFee:       0,
			ProcessingFeeGst:         0,
			FirstEmiDueDate:          "",
			SubventionTag:            nil,
			DiscountedInterestAmount: 0,
			Schedule:                 nil,
		},
	}

	returnURL := req.ReturnURL

	// Call gateway to create order
	gatewayResp, err := s.gateway.CreateOrder(ctx, port.OrderInput{
		MerchantTxnID: merchantTxnID,
		Mobile:        contact.Mobile,
		Email:         contact.Email,
		Amount:        req.Amount,
		Currency:      req.Currency,
		EmiPlans:      emiPlans,
	})
	if err != nil {
		s.idempotency.Fail(ctx, idempotencyKey)
		return nil, err
	}

	// Status: always PENDING on create (gateway may return empty/INITIATED)
	orderStatus := entity.OrderPending
	switch entity.OrderStatus(gatewayResp.Status) {
	case entity.OrderSuccess, entity.OrderFailed, entity.OrderRefunded, entity.OrderExpired, entity.OrderCancelled:
		orderStatus = entity.OrderStatus(gatewayResp.Status)
	case entity.OrderPending:
		orderStatus = entity.OrderPending
	default:
		orderStatus = entity.OrderPending
	}

	lenderOrderIDPtr := (*string)(nil)
	if gatewayResp.LenderOrderID != "" {
		s := gatewayResp.LenderOrderID
		lenderOrderIDPtr = &s
	}

	emiPlanBytes, _ := json.Marshal(map[string]interface{}{"tenure": tenure, "type": emiType})
	order := &entity.Order{
		PaymentID:           paymentID,
		UserID:              req.UserID,
		MerchantID:          merchantID,
		Lender:              "LAZYPAY",
		Amount:              req.Amount,
		Currency:            req.Currency,
		Status:              orderStatus,
		Source:              &source,
		ReturnURL:           &returnURL,
		EMIPlan:             emiPlanBytes,
		LenderOrderID:       lenderOrderIDPtr,
		LenderMerchantTxnID: &merchantTxnID,
		LastErrorCode:       gatewayResp.ErrorCode,
		LastErrorMessage:    gatewayResp.ErrorMessage,
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		s.idempotency.Fail(ctx, idempotencyKey)
		return nil, err
	}

	if gatewayResp.Status == "SUCCESS" || gatewayResp.Status == "PENDING" {
		_, _ = s.contactResolver.RefreshFromSource(ctx, req.UserID, "ORDER")
	}

	mapping := &entity.PaymentMapping{
		PaymentID:           paymentID,
		UserID:              req.UserID,
		Lender:              order.Lender,
		LenderMerchantTxnID: merchantTxnID,
		LenderOrderID:       lenderOrderIDPtr,
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}
	_ = s.mappingRepo.Create(ctx, mapping)

	// Build response — status always "PENDING" on create
	response := &res.OrderResponse{
		PaymentID:     paymentID,
		Status:        "PENDING",
		LenderOrderID: gatewayResp.LenderOrderID,
		RedirectURL:   gatewayResp.RedirectURL,
		Amount:        order.Amount,
		Currency:      order.Currency,
		CreatedAt:     order.CreatedAt,
		ErrorCode:     gatewayResp.ErrorCode,
		ErrorMessage:  gatewayResp.ErrorMessage,
	}

	responseBytes, _ := json.Marshal(response)
	_ = s.idempotency.Complete(ctx, idempotencyKey, responseBytes, lenderOrderIDPtr)

	s.publisher.Publish(ctx, &port.OrderEvent{
		Type:          port.EventOrderCreated,
		PaymentID:     paymentID,
		UserID:        req.UserID,
		MerchantID:    merchantID,
		Lender:        order.Lender,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Status:        "PENDING",
		LenderOrderID: lenderOrderIDPtr,
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

	status := string(order.Status.OrDefault())
	lenderOrderID := ""
	if order.LenderOrderID != nil {
		lenderOrderID = *order.LenderOrderID
	}
	lenderLastStatus := ""
	if order.LenderLastStatus != nil {
		lenderLastStatus = *order.LenderLastStatus
	}
	lenderLastMsg := ""
	if order.LenderLastTxnMessage != nil {
		lenderLastMsg = *order.LenderLastTxnMessage
	}
	lastErrCode := ""
	if order.LastErrorCode != nil {
		lastErrCode = *order.LastErrorCode
	}
	lastErrMsg := ""
	if order.LastErrorMessage != nil {
		lastErrMsg = *order.LastErrorMessage
	}

	// Parse EMI plan from JSONB if present
	var emi *res.EmiDetail
	if len(order.EMIPlan) > 0 {
		var parsed struct {
			Tenure             int     `json:"tenure"`
			EMI                float64 `json:"emi"`
			InterestRate       float64 `json:"interestRate"`
			Principal          float64 `json:"principal"`
			TotalPayableAmount float64 `json:"totalPayableAmount"`
			FirstEmiDueDate    string  `json:"firstEmiDueDate"`
		}
		if err := json.Unmarshal(order.EMIPlan, &parsed); err == nil {
			emi = &res.EmiDetail{
				Tenure:             parsed.Tenure,
				EMI:                parsed.EMI,
				InterestRate:       parsed.InterestRate,
				Principal:          parsed.Principal,
				TotalPayableAmount: parsed.TotalPayableAmount,
				FirstEmiDueDate:    parsed.FirstEmiDueDate,
			}
		}
	}

	return &res.OrderStatusResponse{
		PaymentID:         order.PaymentID,
		Status:            status,
		LenderOrderID:     lenderOrderID,
		Amount:            order.Amount,
		Currency:          order.Currency,
		EmiPlan:           emi,
		LenderLastStatus:  lenderLastStatus,
		LenderLastMessage: lenderLastMsg,
		LastErrorCode:     lastErrCode,
		LastErrorMessage:  lastErrMsg,
		CreatedAt:         order.CreatedAt,
		UpdatedAt:         order.UpdatedAt,
	}, nil
}

// ListByUserID lists orders for a user with optional filters
func (s *OrderService) ListByUserID(ctx context.Context, listReq req.ListOrdersRequest) (*res.OrderListResponse, error) {
	orders, total, err := s.orderRepo.ListByUserID(ctx, listReq.UserID, listReq.MerchantID, listReq.Status, listReq.Page, listReq.PerPage)
	if err != nil {
		return nil, err
	}
	summaries := make([]res.OrderSummary, len(orders))
	for i, o := range orders {
		lenderOrderID := ""
		if o.LenderOrderID != nil {
			lenderOrderID = *o.LenderOrderID
		}
		summaries[i] = res.OrderSummary{
			PaymentID:     o.PaymentID,
			Status:        string(o.Status.OrDefault()),
			LenderOrderID: lenderOrderID,
			Amount:        o.Amount,
			Currency:      o.Currency,
			CreatedAt:     o.CreatedAt,
			UpdatedAt:     o.UpdatedAt,
		}
	}
	if listReq.Page < 1 {
		listReq.Page = 1
	}
	if listReq.PerPage < 1 {
		listReq.PerPage = 20
	}
	return &res.OrderListResponse{
		Orders:  summaries,
		Total:   total,
		Page:    listReq.Page,
		PerPage: listReq.PerPage,
	}, nil
}

// SupportUpdateStatus allows support to override order status (PENDING → FAILED or CANCELLED)
func (s *OrderService) SupportUpdateStatus(ctx context.Context, paymentID string, updateReq req.UpdateOrderStatusRequest) (*entity.Order, error) {
	order, err := s.orderRepo.GetForUpdate(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, sharedErrors.New(sharedErrors.CodeOrderNotFound, 404, "order not found")
	}
	target := entity.OrderStatus(updateReq.Status)
	allowed := order.Status == entity.OrderPending && (target == entity.OrderFailed || target == entity.OrderCancelled)
	if !allowed {
		return nil, sharedErrors.New(sharedErrors.CodeInvalidTransition, 422,
			"cannot move "+string(order.Status)+" → "+string(target)+" via support override")
	}
	reason := updateReq.Reason
	order.Status = target
	order.LastErrorMessage = &reason
	order.LenderLastStatus = ptr("SUPPORT_OVERRIDE")
	order.UpdatedAt = time.Now().UTC()
	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func ptr(s string) *string { return &s }

// ProcessCallback processes an order callback event (from Kafka consumer)
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
