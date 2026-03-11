package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	req "lending-hub-service/internal/domain/onboarding/dto/request"
	"lending-hub-service/internal/domain/onboarding/service"
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/internal/shared/middleware"
	"lending-hub-service/internal/shared/response"
)

// StartOnboardingHandler handles onboarding initiation requests
type StartOnboardingHandler struct {
	service *service.OnboardingService
}

// NewStartOnboardingHandler creates a new StartOnboardingHandler
func NewStartOnboardingHandler(svc *service.OnboardingService) *StartOnboardingHandler {
	return &StartOnboardingHandler{service: svc}
}

// Handle processes start onboarding requests
func (h *StartOnboardingHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var req req.StartOnboardingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "invalid request body: "+err.Error(), requestID)
		c.JSON(status, envelope)
		return
	}

	resp, err := h.service.StartOnboarding(c.Request.Context(), req)
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
