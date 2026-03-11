package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	req "lending-hub-service/internal/domain/profile/dto/request"
	"lending-hub-service/internal/domain/profile/service"
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/internal/shared/middleware"
	"lending-hub-service/internal/shared/response"
)

// EligibilityHandler handles eligibility requests
type EligibilityHandler struct {
	service service.ProfileService
}

// NewEligibilityHandler creates a new EligibilityHandler
func NewEligibilityHandler(svc service.ProfileService) *EligibilityHandler {
	return &EligibilityHandler{service: svc}
}

// Handle processes POST /v1/payin3/eligibility
func (h *EligibilityHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var req req.EligibilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "invalid request body: "+err.Error(), requestID)
		c.JSON(status, envelope)
		return
	}

	resp, err := h.service.CheckEligibility(c.Request.Context(), req)
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
