package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"lending-hub-service/internal/domain/order/service"
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/internal/shared/middleware"
	"lending-hub-service/internal/shared/response"
)

// GetOrderByLoanHandler handles GET /v1/payin3/order/loan/:loanId (internal/support)
type GetOrderByLoanHandler struct {
	service *service.OrderService
}

// NewGetOrderByLoanHandler creates a new GetOrderByLoanHandler
func NewGetOrderByLoanHandler(svc *service.OrderService) *GetOrderByLoanHandler {
	return &GetOrderByLoanHandler{service: svc}
}

// Handle processes get order by loanId requests
func (h *GetOrderByLoanHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	loanID := c.Param("loanId")
	if loanID == "" {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "loanId is required", requestID)
		c.JSON(status, envelope)
		return
	}
	resp, err := h.service.GetOrderStatusByLoanID(c.Request.Context(), loanID)
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
