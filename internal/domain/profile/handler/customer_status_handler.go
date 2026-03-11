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

// CustomerStatusHandler handles customer status requests
type CustomerStatusHandler struct {
	service service.ProfileService
}

// NewCustomerStatusHandler creates a new CustomerStatusHandler
func NewCustomerStatusHandler(svc service.ProfileService) *CustomerStatusHandler {
	return &CustomerStatusHandler{service: svc}
}

// Handle processes customer status requests
func (h *CustomerStatusHandler) Handle(c *gin.Context) {
	requestID := middleware.GetRequestID(c)

	var req req.CustomerStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, envelope := response.Error(http.StatusBadRequest, sharedErrors.CodeInvalidRequest, "invalid request body: "+err.Error(), requestID)
		c.JSON(status, envelope)
		return
	}

	resp, err := h.service.GetCustomerStatus(c.Request.Context(), req)
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
