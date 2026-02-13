package health

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db    *sql.DB
	redis *redis.Client // can be nil if Redis is disabled
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *sql.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

// RegisterRoutes registers health check routes
func (h *HealthHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", h.Health)
	r.GET("/health/ready", h.Ready)
}

// Health — liveness probe: always 200 if process is running
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "payin3-service",
	})
}

// Ready — readiness probe: checks DB and Redis connectivity
func (h *HealthHandler) Ready(c *gin.Context) {
	checks := make(map[string]string)
	healthy := true

	if err := h.db.PingContext(c.Request.Context()); err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		healthy = false
	} else {
		checks["database"] = "ok"
	}

	if h.redis != nil {
		if err := h.redis.Ping(c.Request.Context()).Err(); err != nil {
			checks["redis"] = "unhealthy: " + err.Error()
			healthy = false
		} else {
			checks["redis"] = "ok"
		}
	} else {
		checks["redis"] = "disabled"
	}

	status := http.StatusOK
	if !healthy {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"status": map[bool]string{true: "ok", false: "degraded"}[healthy],
		"checks": checks,
	})
}
