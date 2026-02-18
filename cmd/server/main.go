package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	// Config
	"lending-hub-service/config"
	pg "lending-hub-service/internal/infrastructure/postgres"

	// Infrastructure - Phase 4A
	"lending-hub-service/internal/infrastructure/cache"
	"lending-hub-service/internal/infrastructure/kafka"

	// Infrastructure - Phase 4B (logging + business metrics)
	"lending-hub-service/internal/infrastructure/observability/metrics"
	"lending-hub-service/pkg/idgen"
	baseLogger "lending-hub-service/pkg/logger"
	baseValidator "lending-hub-service/pkg/validator"

	// Infrastructure - Phase 4C
	"lending-hub-service/internal/infrastructure/health"
	"lending-hub-service/internal/infrastructure/middleware"

	// Domain modules
	"lending-hub-service/internal/domain/onboarding"
	onboardingPort "lending-hub-service/internal/domain/onboarding/port"
	onboardingStub "lending-hub-service/internal/domain/onboarding/stub"
	"lending-hub-service/internal/domain/order"
	orderPort "lending-hub-service/internal/domain/order/port"
	orderRepo "lending-hub-service/internal/domain/order/repository"
	orderStub "lending-hub-service/internal/domain/order/stub"
	"lending-hub-service/internal/domain/profile"
	profilePort "lending-hub-service/internal/domain/profile/port"
	profileStub "lending-hub-service/internal/domain/profile/stub"
	"lending-hub-service/internal/domain/refund"
	refundPort "lending-hub-service/internal/domain/refund/port"
	refundStub "lending-hub-service/internal/domain/refund/stub"

	// Adapters
	"lending-hub-service/internal/adapter/lazypay"
	lpConfig "lending-hub-service/internal/adapter/lazypay/config"
)

func main() {
	// ═══════════════════════════════════════
	// 1. Load configuration
	// ═══════════════════════════════════════
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		// Use standard log since logger not initialized yet
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// ═══════════════════════════════════════
	// 2. Initialize logging (Phase 4B)
	// ═══════════════════════════════════════
	logger, err := baseLogger.New(baseLogger.Config{
		Service: "payin3-service",
		Env:     cfg.Env,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// ═══════════════════════════════════════
	// 2.5. Initialize ID generator
	// ═══════════════════════════════════════
	idGen := idgen.New()

	// ═══════════════════════════════════════
	// 2.6. Initialize validator
	// ═══════════════════════════════════════
	validator := baseValidator.New()
	_ = baseValidator.NewHTTPValidator(validator) // Reserved for future handler refactoring

	// ═══════════════════════════════════════
	// 3. Initialize business metrics (Phase 4B + 4B-FIX)
	//    System/provider metrics handled by DD Agent — not here
	// ═══════════════════════════════════════
	var mc metrics.MetricsClient
	ddCfg := metrics.DefaultConfig()
	ddCfg.Enabled = false // TODO: add to config.Config if needed
	if ddCfg.Enabled {
		mc, err = metrics.NewDatadogClient(ddCfg)
		if err != nil {
			logger.Error("failed to init Datadog, falling back to noop",
				baseLogger.ErrorCode(err.Error()))
			mc = metrics.NewNoopClient()
		} else {
			logger.Info("Datadog metrics enabled")
		}
	} else {
		mc = metrics.NewNoopClient()
		logger.Info("Using noop metrics client")
	}
	defer mc.Close()

	// ═══════════════════════════════════════
	// 4. Initialize database
	// ═══════════════════════════════════════
	gormDB, err := pg.NewDB(cfg.DB)
	if err != nil {
		logger.Fatal("failed to connect database", baseLogger.ErrorCode(err.Error()))
	}
	defer func() {
		if err := pg.Close(gormDB); err != nil {
			logger.Error("failed to close database", baseLogger.ErrorCode(err.Error()))
		}
	}()

	// Get underlying sql.DB for health checks
	sqlDB, err := gormDB.DB()
	if err != nil {
		logger.Fatal("failed to get sql.DB", baseLogger.ErrorCode(err.Error()))
	}

	// Apply connection pool settings (already done in NewDB, but ensure)
	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.DB.ConnMaxIdleTime)

	if err := sqlDB.Ping(); err != nil {
		logger.Fatal("database ping failed", baseLogger.ErrorCode(err.Error()))
	}
	logger.Info("database connected")

	// ═══════════════════════════════════════
	// 5. Initialize Redis cache (Phase 4A)
	// ═══════════════════════════════════════
	var redisClient *redis.Client
	var profileCache profilePort.ProfileCache
	if cfg.Redis.Addr != "" {
		redisCfg := cache.RedisConfig{
			Address:      cfg.Redis.Addr,
			Password:     cfg.Redis.Password,
			DB:           cfg.Redis.DB,
			PoolSize:     25,
			MinIdleConns: 5,
			DialTimeout:  cfg.Redis.DialTimeout,
			ReadTimeout:  cfg.Redis.ReadTimeout,
			WriteTimeout: cfg.Redis.WriteTimeout,
		}
		redisCache, err := cache.NewRedisProfileCache(redisCfg)
		if err != nil {
			logger.Error("failed to init Redis, falling back to memory cache",
				baseLogger.ErrorCode(err.Error()))
			profileCache = cache.NewMemoryProfileCache(60 * time.Second) // 60s TTL
			redisClient = nil
		} else {
			// Extract underlying redis.Client for health checks
			redisClient = redisCache.Client()
			profileCache = redisCache
			logger.Info("Using Redis cache")
		}
	} else {
		profileCache = cache.NewMemoryProfileCache(60 * time.Second) // 60s TTL
		logger.Info("Using memory cache (no Redis config)")
	}

	// ═══════════════════════════════════════
	// 6. Initialize Kafka producer (Phase 4A)
	// ═══════════════════════════════════════
	var producer kafka.EventPublisher
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Enabled {
		producerCfg := kafka.ProducerConfig{
			Brokers:         cfg.Kafka.Brokers,
			Async:           true,
			CompressionType: "snappy",
			BatchSize:       100,
			LingerMs:        5,
			Retries:         3,
			RequiredAcks:    "all",
		}
		kafkaProducer, err := kafka.NewProducer(producerCfg)
		if err != nil {
			logger.Error("failed to init Kafka, falling back to noop",
				baseLogger.ErrorCode(err.Error()))
			producer = kafka.NewNoopProducer()
		} else {
			producer = kafkaProducer
			logger.Info("Using Kafka producer")
		}
	} else {
		producer = kafka.NewNoopProducer()
		logger.Info("Using noop producer (no Kafka config)")
	}
	defer producer.Close()

	// Create event publishers for modules
	profileEventPublisher := kafka.NewProfileEventPublisher(producer)
	orderEventPublisher := kafka.NewOrderEventPublisher(producer)

	// ═══════════════════════════════════════
	// 7. Lazypay adapter (Phase 3C)
	// ═══════════════════════════════════════
	var profileGateway profilePort.ProfileGateway
	var onboardingGateway onboardingPort.OnboardingGateway
	var orderGateway orderPort.OrderGateway
	var refundGateway refundPort.RefundGateway

	if cfg.Lazypay.Enabled {
		// Convert config.Config.Lazypay to adapter config
		lpCfg := &lpConfig.LazypayConfig{
			BaseURL:        cfg.Lazypay.BaseURL,
			AccessKey:      cfg.Lazypay.AccessKey,
			SecretKey:      cfg.Lazypay.SecretKey,
			MerchantID:     cfg.Lazypay.MerchantID,
			SubMerchantID:  cfg.Lazypay.SubMerchantID,
			ProfileTimeout: int(cfg.Lazypay.ProfileTimeout.Seconds()),
			PaymentTimeout: int(cfg.Lazypay.PaymentTimeout.Seconds()),
		}
		lazypayClient := lazypay.NewAdapter(lpCfg, logger)
		profileGateway = lazypayClient.ProfileGateway()
		onboardingGateway = lazypayClient.OnboardingGateway()
		orderGateway = lazypayClient.OrderGateway()
		refundGateway = lazypayClient.RefundGateway()
		logger.Info("Using Lazypay adapter")
	} else {
		// Use stubs
		profileGateway = profileStub.NewStubProfileGateway()
		onboardingGateway = onboardingStub.NewStubOnboardingGateway()
		orderGateway = orderStub.NewStubOrderGateway()
		refundGateway = refundStub.NewStubRefundGateway()
		logger.Info("Using stub gateways (Lazypay disabled)")
	}

	// ═══════════════════════════════════════
	// 8. Domain modules
	// ═══════════════════════════════════════
	// Profile module
	profileModule := profile.NewModule(gormDB, profileGateway, profileCache, profileEventPublisher)
	profileUpdater := profileModule.Updater

	// Onboarding module
	onboardingModule := onboarding.NewModule(gormDB, onboardingGateway, profileUpdater, idGen)

	// Order module
	orderModule := order.NewModule(gormDB, orderGateway, profileUpdater, orderEventPublisher, idGen)

	// Refund module
	orderRepo := orderRepo.NewOrderRepository(gormDB)
	refundModule := refund.NewModule(gormDB, refundGateway, orderRepo, profileUpdater)

	// ═══════════════════════════════════════
	// 9. Setup Gin router + middleware
	//    NOTE: No metrics middleware — DD APM auto-instruments Gin
	// ═══════════════════════════════════════
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New() // New() not Default() — we register our own middleware

	// Health endpoints (registered before middleware — no auth/context needed)
	healthHandler := health.NewHealthHandler(sqlDB, redisClient)
	healthHandler.RegisterRoutes(router)

	// Global middleware stack (order matters)
	router.Use(
		middleware.RequestID(idGen),       // 1. Generate/extract request ID
		middleware.Recovery(logger),       // 2. Panic recovery + structured log
		middleware.RequestLogging(logger), // 3. Structured request/response log
		middleware.ContextHeaders(map[string]bool{ // 4. Extract platform context headers
			"/health":       true,
			"/health/ready": true,
		}),
	)

	// API route group
	v1 := router.Group("/v1/payin3")
	{
		profileModule.RegisterRoutes(v1)
		onboardingModule.RegisterRoutes(v1)
		orderModule.RegisterRoutes(v1)
		refundModule.RegisterRoutes(v1)
	}

	// ═══════════════════════════════════════
	// 10. Start HTTP server with graceful shutdown
	// ═══════════════════════════════════════
	port := cfg.HTTP.Port
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("starting server",
			baseLogger.Module("server"),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", baseLogger.ErrorCode(err.Error()))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced shutdown", baseLogger.ErrorCode(err.Error()))
	}

	logger.Info("server stopped")
}
