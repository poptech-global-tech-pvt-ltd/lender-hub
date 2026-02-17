package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"lending-hub-service/internal/domain/refund/service"
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/internal/shared/middleware"
	"lending-hub-service/internal/shared/response"
)

// GetRefundHandler handles refund status requests
type GetRefundHandler struct {
	service *service.RefundService
}

// NewGetRefundHandler creates a new GetRefundHandler
func NewGetRefundHandler(svc *service.RefundService) *GetRefundHandler {
	return &GetRefundHandler{service: svc}
}

// Handle processes get refund status requests
func (h *GetRefundHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	refundID := c.Param("refundId")
	if refundID == "" {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "refundId is required", requestID)
		c.JSON(status, envelope)
		return
	}

	resp, err := h.service.GetRefundStatus(c.Request.Context(), refundID)
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
