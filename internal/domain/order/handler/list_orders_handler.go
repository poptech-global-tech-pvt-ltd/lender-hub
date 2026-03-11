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

// ListOrdersHandler handles GET /v1/payin3/orders
type ListOrdersHandler struct {
	service *service.OrderService
}

// NewListOrdersHandler creates a new ListOrdersHandler
func NewListOrdersHandler(svc *service.OrderService) *ListOrdersHandler {
	return &ListOrdersHandler{service: svc}
}

// Handle processes list orders requests
func (h *ListOrdersHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var listReq req.ListOrdersRequest
	if err := c.ShouldBindQuery(&listReq); err != nil {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "invalid query: "+err.Error(), requestID)
		c.JSON(status, envelope)
		return
	}

	resp, err := h.service.ListByUserID(c.Request.Context(), listReq)
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
