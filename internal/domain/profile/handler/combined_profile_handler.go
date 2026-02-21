package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"lending-hub-service/internal/domain/profile/service"
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/internal/shared/middleware"
	"lending-hub-service/internal/shared/response"
)

// CombinedProfileHandler handles GET /v1/payin3/profile/:userId
type CombinedProfileHandler struct {
	service service.ProfileService
}

// NewCombinedProfileHandler creates a new CombinedProfileHandler
func NewCombinedProfileHandler(svc service.ProfileService) *CombinedProfileHandler {
	return &CombinedProfileHandler{service: svc}
}

// Handle processes GET /v1/payin3/profile/:userId
func (h *CombinedProfileHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	userID := c.Param("userId")
	if userID == "" {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "userId is required", requestID)
		c.JSON(status, envelope)
		return
	}

	source := c.Query("source")
	if source == "" {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "source query param is required", requestID)
		c.JSON(status, envelope)
		return
	}

	currency := c.DefaultQuery("currency", "INR")

	var amount float64
	if amtStr := c.Query("amount"); amtStr != "" {
		parsed, err := strconv.ParseFloat(amtStr, 64)
		if err != nil || parsed <= 0 {
			status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "amount must be a positive number", requestID)
			c.JSON(status, envelope)
			return
		}
		amount = parsed
	}

	resp, err := h.service.GetCombinedProfile(c.Request.Context(), userID, amount, currency, source)
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
