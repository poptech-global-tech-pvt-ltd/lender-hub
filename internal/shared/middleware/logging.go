package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logging logs each HTTP request with method, path, status, duration
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		requestID := c.GetString("request_id")

		// Process request
		c.Next()

		// Log after request completes
		duration := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()

		log.Printf(
			"[%s] %s %s %d %s from %s",
			requestID,
			method,
			path,
			status,
			duration,
			clientIP,
		)

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				log.Printf("[%s] ERROR: %v", requestID, e.Err)
			}
		}
	}
}
