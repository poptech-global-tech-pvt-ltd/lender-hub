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

// SupportOrderByLoanHandler handles PATCH /v1/payin3/order/loan/:loanId/status
type SupportOrderByLoanHandler struct {
	service           *service.OrderService
	internalAPIToken  string
}

// NewSupportOrderByLoanHandler creates a new SupportOrderByLoanHandler
func NewSupportOrderByLoanHandler(svc *service.OrderService, internalAPIToken string) *SupportOrderByLoanHandler {
	return &SupportOrderByLoanHandler{service: svc, internalAPIToken: internalAPIToken}
}

// Handle processes support status override by loanId
func (h *SupportOrderByLoanHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	token := c.GetHeader(internalTokenHeader)
	if token == "" || token != h.internalAPIToken {
		status, envelope := response.Error(http.StatusUnauthorized, "UNAUTHORIZED", "invalid or missing internal token", requestID)
		c.JSON(status, envelope)
		return
	}

	loanID := c.Param("loanId")
	if loanID == "" {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "loanId is required", requestID)
		c.JSON(status, envelope)
		return
	}

	var updateReq req.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "invalid body: "+err.Error(), requestID)
		c.JSON(status, envelope)
		return
	}

	order, err := h.service.SupportUpdateStatusByLoanID(c.Request.Context(), loanID, updateReq)
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

	loanIDResp := ""
	if order.LenderMerchantTxnID != nil {
		loanIDResp = *order.LenderMerchantTxnID
	}
	status, envelope := response.OK(gin.H{
		"loanId":    loanIDResp,
		"paymentId": order.PaymentID,
		"status":    string(order.Status),
		"updatedAt": order.UpdatedAt,
	})
	c.JSON(status, envelope)
}
