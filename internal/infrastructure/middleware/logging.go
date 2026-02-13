package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"lending-hub-service/internal/infrastructure/observability/logging"
)

// RequestLogging logs structured request/response information
func RequestLogging(logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		requestID, _ := c.Get("requestId")
		requestIDStr := ""
		if requestID != nil {
			requestIDStr = requestID.(string)
		}

		fields := []zap.Field{
			logging.RequestID(requestIDStr),
			zap.String("method", method),
			zap.String("path", path),
			logging.HTTPStatus(status),
			logging.DurationMs(duration.Milliseconds()),
			zap.Int("bodySize", c.Writer.Size()),
		}

		if uid, exists := c.Get("userId"); exists && uid != nil {
			fields = append(fields, logging.UserID(uid.(string)))
		}

		switch {
		case status >= 500:
			logger.Error("request completed", fields...)
		case status >= 400:
			logger.Warn("request completed", fields...)
		default:
			logger.Info("request completed", fields...)
		}
	}
}
