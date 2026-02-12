package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lending-hub-service/config"
	pg "lending-hub-service/internal/infrastructure/postgres"
	"lending-hub-service/internal/shared/middleware"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting lending-hub-service in %s environment", cfg.Env)

	// Initialize database
	db, err := pg.NewDB(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := pg.Close(db); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Verify database connection
	if err := pg.HealthCheck(db); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}
	log.Println("Database connection established")

	// Set Gin mode based on environment
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	router := gin.New()

	// Register global middleware
	router.Use(
		middleware.RequestID(),
		middleware.Logging(),
		middleware.Recovery(),
	)

	// Register routes
	registerRoutes(router, db)

	// Create HTTP server
	addr := ":" + cfg.HTTP.Port
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}

// registerRoutes sets up all HTTP routes
func registerRoutes(router *gin.Engine, db *gorm.DB) {
	// Health check endpoint
	router.GET("/health", healthHandler(db))

	// Readiness check endpoint
	router.GET("/ready", readyHandler(db))

	// API version group (future module routes will be registered here)
	v1 := router.Group("/v1")
	{
		// Placeholder for future routes
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
				"service": "lending-hub-service",
			})
		})
	}
}

// healthHandler returns basic health status
func healthHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check database health
		if err := pg.HealthCheck(db); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"checks": gin.H{
					"database": gin.H{
						"status": "down",
						"error":  err.Error(),
					},
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"checks": gin.H{
				"database": gin.H{
					"status": "up",
				},
			},
		})
	}
}

// readyHandler returns readiness status with detailed metrics
func readyHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check database health
		if err := pg.HealthCheck(db); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"ready": false,
				"error": err.Error(),
			})
			return
		}

		// Get database stats
		stats, err := pg.Stats(db)
		if err != nil {
			log.Printf("Failed to get DB stats: %v", err)
			stats = gin.H{"error": "stats unavailable"}
		}

		c.JSON(http.StatusOK, gin.H{
			"ready":    true,
			"database": stats,
		})
	}
}
