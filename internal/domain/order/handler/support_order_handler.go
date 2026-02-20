package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	req "lending-hub-service/internal/domain/order/dto/request"
	"lending-hub-service/internal/domain/order/service"
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/internal/shared/middleware"
	"lending-hub-service/internal/shared/response"
)

const internalTokenHeader = "X-Internal-Token"

// SupportOrderHandler handles PATCH /v1/payin3/order/:paymentId/status
type SupportOrderHandler struct {
	service        *service.OrderService
	internalAPIToken string
}

// NewSupportOrderHandler creates a new SupportOrderHandler
func NewSupportOrderHandler(svc *service.OrderService, internalAPIToken string) *SupportOrderHandler {
	return &SupportOrderHandler{service: svc, internalAPIToken: internalAPIToken}
}

// Handle processes support status override requests
func (h *SupportOrderHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	token := c.GetHeader(internalTokenHeader)
	if token == "" || token != h.internalAPIToken {
		status, envelope := response.Error(http.StatusUnauthorized, "UNAUTHORIZED", "invalid or missing internal token", requestID)
		c.JSON(status, envelope)
		return
	}

	paymentID := c.Param("paymentId")
	if paymentID == "" {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "paymentId is required", requestID)
		c.JSON(status, envelope)
		return
	}

	var updateReq req.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "invalid body: "+err.Error(), requestID)
		c.JSON(status, envelope)
		return
	}

	order, err := h.service.SupportUpdateStatus(c.Request.Context(), paymentID, updateReq)
	if err != nil {
		if de, ok := err.(*sharedErrors.DomainError); ok {
			status, envelope := response.Error(de.Status, de.Code, de.Message, requestID)
			c.JSON(status, envelope)
			return
		}
		status, envelope := response.Error(http.StatusInternalServerError, sharedErrors.CodeInternalError, "internal server error", requestID)
		c.JSON(status, envelope)
		return
	}

	status, envelope := response.OK(gin.H{
		"paymentId": order.PaymentID,
		"status":    string(order.Status),
		"updatedAt": order.UpdatedAt,
	})
	c.JSON(status, envelope)
}
