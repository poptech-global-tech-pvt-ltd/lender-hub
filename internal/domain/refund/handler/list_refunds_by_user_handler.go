package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	res "lending-hub-service/internal/domain/refund/dto/response"
	"lending-hub-service/internal/domain/refund/service"
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/internal/shared/middleware"
	"lending-hub-service/internal/shared/response"
)

// ListRefundsByUserHandler handles GET /refunds/user?userId=...&page=1&perPage=20
type ListRefundsByUserHandler struct {
	service *service.RefundService
}

// NewListRefundsByUserHandler creates a new ListRefundsByUserHandler
func NewListRefundsByUserHandler(svc *service.RefundService) *ListRefundsByUserHandler {
	return &ListRefundsByUserHandler{service: svc}
}

// Handle lists all refunds for a user
func (h *ListRefundsByUserHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)
	userID := c.Query("userId")
	if userID == "" {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "userId is required", requestID)
		c.JSON(status, envelope)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "20"))

	refunds, total, err := h.service.ListByUserID(c.Request.Context(), userID, page, perPage)
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
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}
	status, envelope := response.OK(resp)
	c.JSON(status, envelope)
}
