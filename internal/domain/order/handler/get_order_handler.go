package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"lending-hub-service/internal/domain/order/service"
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/internal/shared/middleware"
	"lending-hub-service/internal/shared/response"
)

// GetOrderHandler handles order status requests
type GetOrderHandler struct {
	service *service.OrderService
}

// NewGetOrderHandler creates a new GetOrderHandler
func NewGetOrderHandler(svc *service.OrderService) *GetOrderHandler {
	return &GetOrderHandler{service: svc}
}

// Handle processes get order status requests
func (h *GetOrderHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	paymentID := c.Param("paymentId")
	if paymentID == "" {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "paymentId is required", requestID)
		c.JSON(status, envelope)
		return
	}

	resp, err := h.service.GetOrderStatus(c.Request.Context(), paymentID)
	if err != nil {
		if de, ok := err.(*sharedErrors.DomainError); ok {
			status, envelope := response.Error(de.Status, de.Code, de.Message, requestID)
			c.JSON(status, envelope)
			return
		}
		// Unknown error
		status, envelope := response.Error(http.StatusInternalServerError, sharedErrors.CodeInternalError, "internal server error", requestID)
		c.JSON(status, envelope)
		return
	}

	status, envelope := response.OK(resp)
	c.JSON(status, envelope)
}
