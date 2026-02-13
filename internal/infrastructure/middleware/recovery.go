package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"lending-hub-service/internal/infrastructure/observability/logging"
)

// Recovery catches panics, logs full stack trace, returns 500 with canonical error envelope
func Recovery(logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := string(debug.Stack())
				requestID, _ := c.Get("requestId")
				requestIDStr := ""
				if requestID != nil {
					requestIDStr = requestID.(string)
				}

				logger.Error("panic recovered",
					logging.RequestID(requestIDStr),
					zap.Any("error", err),
					zap.String("stack", stack),
					zap.String("path", c.Request.URL.Path),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"data":    nil,
					"error": gin.H{
						"code":       "PAYIN3_INTERNAL_ERROR",
						"message":    "An internal error occurred",
						"statusCode": 500,
						"retryable":  true,
					},
				})
			}
		}()
		c.Next()
	}
}
