package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// Recovery recovers from panics and returns 500 error
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				requestID := c.GetString("request_id")
				stack := string(debug.Stack())

				log.Printf(
					"[%s] PANIC RECOVERED: %v\nStack trace:\n%s",
					requestID,
					r,
					stack,
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":      "internal_server_error",
					"message":    "An unexpected error occurred",
					"request_id": requestID,
				})
			}
		}()

		c.Next()
	}
}
