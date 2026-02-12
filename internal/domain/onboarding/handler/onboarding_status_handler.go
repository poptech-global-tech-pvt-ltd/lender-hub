package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"lending-hub-service/internal/domain/onboarding/service"
	sharedErrors "lending-hub-service/internal/shared/errors"
	"lending-hub-service/internal/shared/middleware"
	"lending-hub-service/internal/shared/response"
)

// OnboardingStatusHandler handles onboarding status requests
type OnboardingStatusHandler struct {
	service *service.OnboardingService
}

// NewOnboardingStatusHandler creates a new OnboardingStatusHandler
func NewOnboardingStatusHandler(svc *service.OnboardingService) *OnboardingStatusHandler {
	return &OnboardingStatusHandler{service: svc}
}

// Handle processes onboarding status requests
func (h *OnboardingStatusHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	userID := c.Query("userId")
	onboardingID := c.Query("onboardingId")
	merchantID := c.Query("merchantId")

	if userID == "" {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "userId is required", requestID)
		c.JSON(status, envelope)
		return
	}

	if merchantID == "" {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "merchantId is required", requestID)
		c.JSON(status, envelope)
		return
	}

	resp, err := h.service.GetStatus(c.Request.Context(), userID, onboardingID, merchantID)
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
