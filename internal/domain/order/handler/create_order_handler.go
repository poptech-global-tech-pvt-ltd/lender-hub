package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	req "lending-hub-service/internal/domain/order/dto/request"
	"lending-hub-service/internal/domain/order/service"
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/internal/shared/middleware"
	"lending-hub-service/internal/shared/response"
	baseLogger "lending-hub-service/pkg/logger"
)

// CreateOrderHandler handles order creation requests
type CreateOrderHandler struct {
	service *service.OrderService
	logger  *baseLogger.Logger
}

// NewCreateOrderHandler creates a new CreateOrderHandler
func NewCreateOrderHandler(svc *service.OrderService, logger *baseLogger.Logger) *CreateOrderHandler {
	return &CreateOrderHandler{service: svc, logger: logger}
}

// Handle processes create order requests
func (h *CreateOrderHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var req req.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "invalid request body: "+err.Error(), requestID)
		c.JSON(status, envelope)
		return
	}

	resp, err := h.service.CreateOrder(c.Request.Context(), req)
	if err != nil {
		if de, ok := err.(*sharedErrors.DomainError); ok {
			status, envelope := response.Error(de.Status, de.Code, de.Message, requestID)
			c.JSON(status, envelope)
			return
		}
		// Unknown error — log so we can see Lazypay/DB/idempotency failures
		if h.logger != nil {
			h.logger.Error("create order failed", baseLogger.Module("order"), zap.String("requestId", requestID), zap.String("paymentId", req.PaymentID), zap.Error(err))
		}
		status, envelope := response.Error(http.StatusInternalServerError, sharedErrors.CodeInternalError, "internal server error", requestID)
		c.JSON(status, envelope)
		return
	}

	status, envelope := response.OK(resp)
	c.JSON(status, envelope)
}
