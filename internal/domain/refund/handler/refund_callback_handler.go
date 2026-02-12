package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	req "lending-hub-service/internal/domain/refund/dto/request"
	"lending-hub-service/internal/domain/refund/service"
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/internal/shared/middleware"
	"lending-hub-service/internal/shared/response"
)

// RefundCallbackHandler handles refund callback requests
type RefundCallbackHandler struct {
	service *service.RefundService
}

// NewRefundCallbackHandler creates a new RefundCallbackHandler
func NewRefundCallbackHandler(svc *service.RefundService) *RefundCallbackHandler {
	return &RefundCallbackHandler{service: svc}
}

// Handle processes refund callback requests
func (h *RefundCallbackHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var req req.RefundCallbackRequest
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
