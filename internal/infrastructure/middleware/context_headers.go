package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	reqctx "lending-hub-service/internal/shared/context"
)

const (
	HeaderPlatform = "x-platform"
	HeaderDeviceID = "x-device-id"
	HeaderUserIP   = "x-user-ip"
)

// ContextHeaders extracts platform context from headers
// skipPaths: paths that do not require context headers (e.g., /health)
func ContextHeaders(skipPaths map[string]bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		platform := c.GetHeader(HeaderPlatform)
		deviceID := c.GetHeader(HeaderDeviceID)
		userIP := c.GetHeader(HeaderUserIP)

		if platform == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"data":    nil,
				"error": gin.H{
					"code":       "PAYIN3_INVALID_CONTEXT",
					"message":    "Missing required header: x-platform",
					"statusCode": 400,
					"retryable":  false,
				},
			})
			return
		}

		requestID, _ := c.Get("requestId")
		requestIDStr := ""
		if requestID != nil {
			requestIDStr = requestID.(string)
		}
		rc := &reqctx.RequestContext{
			RequestID: requestIDStr,
			Platform:  platform,
			DeviceID:  deviceID,
			UserIP:    userIP,
		}

		c.Set("requestContext", rc)

		ctx := reqctx.WithRequestContext(c.Request.Context(), rc)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
