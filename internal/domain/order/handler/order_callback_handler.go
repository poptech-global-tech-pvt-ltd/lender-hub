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

// OrderCallbackHandler handles order callback requests
type OrderCallbackHandler struct {
	service *service.OrderService
}

// NewOrderCallbackHandler creates a new OrderCallbackHandler
func NewOrderCallbackHandler(svc *service.OrderService) *OrderCallbackHandler {
	return &OrderCallbackHandler{service: svc}
}

// Handle processes order callback requests
func (h *OrderCallbackHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var req req.OrderCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "invalid request body: "+err.Error(), requestID)
		c.JSON(status, envelope)
		return
	}

	err := h.service.ProcessCallback(c.Request.Context(), req)
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

	// Return accepted response
	status, envelope := response.OK(map[string]bool{"accepted": true})
	c.JSON(status, envelope)
}
