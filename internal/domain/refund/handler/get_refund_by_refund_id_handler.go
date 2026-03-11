package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"lending-hub-service/internal/domain/refund/service"
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/internal/shared/middleware"
	"lending-hub-service/internal/shared/response"
)

// GetRefundByRefundIDHandler handles GET /refund/loan/:refundId (by our generated refundId)
type GetRefundByRefundIDHandler struct {
	service *service.RefundService
}

// NewGetRefundByRefundIDHandler creates a new GetRefundByRefundIDHandler
func NewGetRefundByRefundIDHandler(svc *service.RefundService) *GetRefundByRefundIDHandler {
	return &GetRefundByRefundIDHandler{service: svc}
}

// Handle processes get by refundId
func (h *GetRefundByRefundIDHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	refundID := c.Param("refundId")
	if refundID == "" {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "refundId is required", requestID)
		c.JSON(status, envelope)
		return
	}

	refund, err := h.service.GetByRefundID(c.Request.Context(), refundID)
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
