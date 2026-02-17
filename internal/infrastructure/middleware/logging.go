package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	baseLogger "lending-hub-service/pkg/logger"
)

// RequestLogging logs structured request/response information
func RequestLogging(logger *baseLogger.Logger) gin.HandlerFunc {
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
			baseLogger.RequestID(requestIDStr),
			zap.String("method", method),
			zap.String("path", path),
			baseLogger.HTTPStatus(status),
			baseLogger.DurationMs(duration.Milliseconds()),
			zap.Int("bodySize", c.Writer.Size()),
		}

		if uid, exists := c.Get("userId"); exists && uid != nil {
			fields = append(fields, baseLogger.UserID(uid.(string)))
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
