package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"lending-hub-service/internal/domain/order/service"
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/internal/shared/middleware"
	"lending-hub-service/internal/shared/response"
)

// GetOrderByReconHandler handles GET /v1/payin3/order/recon/:lenderOrderId (recon, no enquiry)
type GetOrderByReconHandler struct {
	service *service.OrderService
}

// NewGetOrderByReconHandler creates a new GetOrderByReconHandler
func NewGetOrderByReconHandler(svc *service.OrderService) *GetOrderByReconHandler {
	return &GetOrderByReconHandler{service: svc}
}

// Handle processes get order by Lazypay's lenderOrderId (recon)
func (h *GetOrderByReconHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	lenderOrderID := c.Param("lenderOrderId")
	if lenderOrderID == "" {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "lenderOrderId is required", requestID)
		c.JSON(status, envelope)
		return
	}
	resp, err := h.service.GetOrderStatusByLenderOrderID(c.Request.Context(), lenderOrderID)
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
