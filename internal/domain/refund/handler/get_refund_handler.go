package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"lending-hub-service/internal/domain/refund/service"
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/internal/shared/middleware"
	"lending-hub-service/internal/shared/response"
)

// GetRefundHandler handles GET /refund/:paymentRefundId (by POP's paymentRefundId)
type GetRefundHandler struct {
	service *service.RefundService
}

// NewGetRefundHandler creates a new GetRefundHandler
func NewGetRefundHandler(svc *service.RefundService) *GetRefundHandler {
	return &GetRefundHandler{service: svc}
}

// Handle processes get by paymentRefundId
func (h *GetRefundHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	paymentRefundID := c.Param("paymentRefundId")
	if paymentRefundID == "" {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "paymentRefundId is required", requestID)
		c.JSON(status, envelope)
		return
	}

	refund, err := h.service.GetByPaymentRefundID(c.Request.Context(), paymentRefundID)
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

	resp := service.MapRefundToResponse(refund)
	status, envelope := response.OK(resp)
	c.JSON(status, envelope)
}
