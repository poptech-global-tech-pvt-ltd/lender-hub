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

// CreateRefundHandler handles refund creation requests
type CreateRefundHandler struct {
	service *service.RefundService
}

// NewCreateRefundHandler creates a new CreateRefundHandler
func NewCreateRefundHandler(svc *service.RefundService) *CreateRefundHandler {
	return &CreateRefundHandler{service: svc}
}

// Handle processes create refund requests
func (h *CreateRefundHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var req req.CreateRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "invalid request body: "+err.Error(), requestID)
		c.JSON(status, envelope)
		return
	}

	resp, err := h.service.CreateRefund(c.Request.Context(), req)
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
