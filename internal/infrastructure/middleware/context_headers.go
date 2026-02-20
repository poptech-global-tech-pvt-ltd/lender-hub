package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	reqctx "lending-hub-service/internal/shared/context"
)

const (
	HeaderPlatform = "x-platform"
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
		userIP := c.GetHeader(HeaderUserIP)
		deviceID := c.GetHeader("x-device-id") // Optional, not validated

		missing := make([]string, 0, 2)
		if platform == "" {
			missing = append(missing, "x-platform")
		}
		if userIP == "" {
			missing = append(missing, "x-user-ip")
		}

		if len(missing) > 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"data":    nil,
				"error": gin.H{
					"code":       "PAYIN3_INVALID_CONTEXT",
					"message":    "Missing required context headers: " + joinMissing(missing),
					"statusCode": 400,
					"retryable":  false,
				},
			})
			return
		}

		requestID, _ := c.Get("request_id") // same key as shared/middleware.RequestID
		requestIDStr := ""
		if requestID != nil {
			requestIDStr, _ = requestID.(string)
		}
		rc := &reqctx.RequestContext{
			RequestID: requestIDStr,
			Platform:  platform,
			DeviceID:  deviceID, // Optional, may be empty
			UserIP:    userIP,
		}

		c.Set("requestContext", rc)

		ctx := reqctx.WithRequestContext(c.Request.Context(), rc)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// joinMissing is a simple helper to join missing header names.
func joinMissing(headers []string) string {
	if len(headers) == 0 {
		return ""
	}
	if len(headers) == 1 {
		return headers[0]
	}
	// simple comma+space join
	out := headers[0]
	for i := 1; i < len(headers); i++ {
		out += ", " + headers[i]
	}
	return out
}
