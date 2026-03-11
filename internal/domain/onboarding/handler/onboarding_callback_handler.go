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

// OnboardingCallbackHandler handles onboarding callback requests
type OnboardingCallbackHandler struct {
	service *service.OnboardingService
}

// NewOnboardingCallbackHandler creates a new OnboardingCallbackHandler
func NewOnboardingCallbackHandler(svc *service.OnboardingService) *OnboardingCallbackHandler {
	return &OnboardingCallbackHandler{service: svc}
}

// Handle processes onboarding callback requests
func (h *OnboardingCallbackHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var req req.OnboardingCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "invalid request body: "+err.Error(), requestID)
		c.JSON(status, envelope)
		return
	}

	err := h.service.ProcessCallback(c.Request.Context(), req)
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

	// Return accepted response
	status, envelope := response.OK(map[string]bool{"accepted": true})
	c.JSON(status, envelope)
}
