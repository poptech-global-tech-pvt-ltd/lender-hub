package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"lending-hub-service/internal/domain/order/service"
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/internal/shared/middleware"
	"lending-hub-service/internal/shared/response"
)

// OrderSummaryHandler handles GET /v1/payin3/order/:paymentId/summary
type OrderSummaryHandler struct {
	svc *service.OrderSummaryService
}

// NewOrderSummaryHandler creates a new OrderSummaryHandler
func NewOrderSummaryHandler(svc *service.OrderSummaryService) *OrderSummaryHandler {
	return &OrderSummaryHandler{svc: svc}
}

// Handle processes the order summary request
func (h *OrderSummaryHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	paymentID := c.Param("paymentId")
	if paymentID == "" {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "paymentId is required", requestID)
		c.JSON(status, envelope)
		return
	}

	resp, err := h.svc.GetSummary(c.Request.Context(), paymentID)
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

	status, envelope := response.OK(resp)
	c.JSON(status, envelope)
}
