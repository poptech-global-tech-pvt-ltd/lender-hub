package middleware

import (
	"github.com/gin-gonic/gin"

	"lending-hub-service/pkg/idgen"
)

const (
	// RequestIDHeader is the HTTP header name for request ID
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey is the context key for request ID
	RequestIDKey = "request_id"
)

// RequestID injects or preserves request ID for tracing
func RequestID(idgen *idgen.Generator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request already has ID
		requestID := c.Request.Header.Get(RequestIDHeader)

		// Generate new ID if missing
		if requestID == "" {
			requestID = idgen.RequestID()
		}

		// Store in context for handlers
		c.Set(RequestIDKey, requestID)

		// Add to response headers
		c.Writer.Header().Set(RequestIDHeader, requestID)

		c.Next()
	}
}

// GetRequestID retrieves request ID from context
func GetRequestID(c *gin.Context) string {
	if rid, exists := c.Get(RequestIDKey); exists {
		if ridStr, ok := rid.(string); ok {
			return ridStr
		}
	}
	return ""
}
