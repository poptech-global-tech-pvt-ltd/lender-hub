package middleware

import (
	"github.com/gin-gonic/gin"

	"lending-hub-service/pkg/idgen"
)

const HeaderRequestID = "X-Request-ID"

// RequestID generates or extracts X-Request-ID for every request
func RequestID(idgen *idgen.Generator) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(HeaderRequestID)
		if requestID == "" {
			requestID = idgen.RequestID()
		}

		c.Set("requestId", requestID)
		c.Header(HeaderRequestID, requestID)

		c.Next()
	}
}
