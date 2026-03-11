package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	res "lending-hub-service/internal/domain/refund/dto/response"
	"lending-hub-service/internal/domain/refund/service"
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/internal/shared/middleware"
	"lending-hub-service/internal/shared/response"
)

// ListRefundsForOrderHandler handles GET /refunds?paymentId=...
type ListRefundsForOrderHandler struct {
	service *service.RefundService
}

// NewListRefundsForOrderHandler creates a new ListRefundsForOrderHandler
func NewListRefundsForOrderHandler(svc *service.RefundService) *ListRefundsForOrderHandler {
	return &ListRefundsForOrderHandler{service: svc}
}

// Handle lists all refunds for an order
func (h *ListRefundsForOrderHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	paymentID := c.Query("paymentId")
	if paymentID == "" {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "paymentId is required", requestID)
		c.JSON(status, envelope)
		return
	}

	refunds, err := h.service.ListByPaymentID(c.Request.Context(), paymentID)
	if err != nil {
		status, envelope := response.Error(http.StatusInternalServerError, sharedErrors.CodeInternalError, "failed to list refunds", requestID)
		c.JSON(status, envelope)
		return
	}

	summaries := make([]res.RefundSummary, len(refunds))
	for i, r := range refunds {
		summaries[i] = res.RefundSummary{
			RefundID:            r.RefundID,
			PaymentRefundID:     r.PaymentRefundID,
			ProviderRefundTxnID: "",
			PaymentID:           r.PaymentID,
			LoanID:              r.LoanID,
			Status:              string(r.Status.OrDefault()),
			Amount:              r.Amount,
			Currency:            r.Currency,
			CreatedAt:           r.CreatedAt,
			UpdatedAt:           r.UpdatedAt,
		}
		if r.ProviderRefundTxnID != nil {
			summaries[i].ProviderRefundTxnID = *r.ProviderRefundTxnID
		}
	}

	resp := res.RefundListResponse{
		Refunds: summaries,
		Total:   len(summaries),
		Page:    1,
		PerPage: len(summaries),
	}
	status, envelope := response.OK(resp)
	c.JSON(status, envelope)
}
