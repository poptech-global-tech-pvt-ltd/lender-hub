package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
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
	kafkaConsumer "lending-hub-service/internal/infrastructure/kafka/consumer"

	// Infrastructure - Phase 4B (logging + business metrics)
	"lending-hub-service/internal/infrastructure/observability/metrics"
	"lending-hub-service/internal/infrastructure/userprofile"
	"lending-hub-service/pkg/idgen"
	baseLogger "lending-hub-service/pkg/logger"
	baseValidator "lending-hub-service/pkg/validator"

	// Infrastructure - Phase 4C
	"lending-hub-service/internal/infrastructure/health"
	"lending-hub-service/internal/infrastructure/middleware"

	// Domain modules
	"lending-hub-service/internal/domain/onboarding"
	onboardingPort "lending-hub-service/internal/domain/onboarding/port"
	"lending-hub-service/internal/domain/order"
	orderPort "lending-hub-service/internal/domain/order/port"
	orderRepo "lending-hub-service/internal/domain/order/repository"
	"lending-hub-service/internal/domain/profile"
	profilePort "lending-hub-service/internal/domain/profile/port"
	profileRepo "lending-hub-service/internal/domain/profile/repository"
	profileService "lending-hub-service/internal/domain/profile/service"
	"lending-hub-service/internal/domain/refund"
	refundPort "lending-hub-service/internal/domain/refund/port"

	// Adapters
	"lending-hub-service/internal/adapter/contact"
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
	// 6. Initialize Kafka (producers + event publishers)
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

	profileEventPublisher := kafka.NewProfileEventPublisher(producer)

	var orderEventPublisher orderPort.OrderEventPublisher
	var refundEventPublisher refundPort.RefundEventPublisher
	var orderKafkaPublisher *kafka.OrderEventPublisher
	var refundKafkaPublisher *kafka.RefundEventPublisher
	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Enabled {
		orderKafkaPublisher = kafka.NewOrderEventPublisher(cfg, logger)
		refundKafkaPublisher = kafka.NewRefundEventPublisher(cfg, logger)
		orderEventPublisher = orderKafkaPublisher
		refundEventPublisher = refundKafkaPublisher
		defer func() {
			_ = orderKafkaPublisher.Close()
			_ = refundKafkaPublisher.Close()
		}()
	} else {
		orderEventPublisher = order.NewStubOrderEventPublisher()
		refundEventPublisher = refund.NewStubRefundEventPublisher()
	}

	// ═══════════════════════════════════════
	// 6.5. User Contact Resolver
	// ═══════════════════════════════════════
	// User Profile Service client
	profileServiceClient := userprofile.NewClient(
		cfg.UserProfileService.BaseURL,
		cfg.UserProfileService.Timeout,
	)

	// User Contact Repository
	userContactRepo := profileRepo.NewUserContactRepository(sqlDB)

	// User Contact Resolver
	contactResolver := profileService.NewUserContactResolver(
		userContactRepo,
		profileServiceClient,
		logger,
	)

	// ═══════════════════════════════════════
	// 7. Lazypay adapter (Phase 3C)
	// ═══════════════════════════════════════
	var profileGateway profilePort.ProfileGateway
	var onboardingGateway onboardingPort.OnboardingGateway
	var orderGateway orderPort.OrderGateway
	var refundGateway refundPort.RefundGateway

	// Always use Lazypay adapter (stubs only in tests)
	lpCfg := &lpConfig.LazypayConfig{
		BaseURL:        cfg.Lazypay.BaseURL,
		AccessKey:      cfg.Lazypay.AccessKey,
		SecretKey:      cfg.Lazypay.SecretKey,
		MerchantID:     cfg.Lazypay.MerchantID,
		SubMerchantID:  cfg.Lazypay.SubMerchantID,
		ReturnURL:      cfg.Lazypay.ReturnURL,
		ProfileTimeout: int(cfg.Lazypay.ProfileTimeout.Seconds()),
		PaymentTimeout: int(cfg.Lazypay.PaymentTimeout.Seconds()),
	}
	lazypayClient := lazypay.NewAdapter(lpCfg, logger, idGen)
	profileGateway = lazypayClient.ProfileGateway()
	onboardingGateway = lazypayClient.OnboardingGateway()
	orderGateway = lazypayClient.OrderGateway()
	refundGateway = lazypayClient.RefundGateway()
	logger.Info("Using Lazypay adapter")

	// ═══════════════════════════════════════
	// 8. Domain modules
	// ═══════════════════════════════════════
	// Profile module (cache no longer used — persistence to lender_user)
	_ = profileCache
	profileModule := profile.NewModule(gormDB, profileGateway, profileEventPublisher, contactResolver, profileServiceClient, logger)
	profileUpdater := profileModule.Updater

	// Onboarding module (profileUpdater satisfies port.ProfileUpdater; adapter for ContactResolver)
	onboardingContactResolver := contact.NewOnboardingContactAdapter(contactResolver)
	onboardingModule := onboarding.NewModule(gormDB, onboardingGateway, profileUpdater, idGen, onboardingContactResolver, logger)

	// Order module (merchantID for lender_payment_state.merchant_id NOT NULL)
	orderMerchantID := cfg.Lazypay.MerchantID
	if orderMerchantID == "" {
		orderMerchantID = cfg.Lazypay.SubMerchantID
	}
	orderContactResolver := contact.NewOrderContactAdapter(contactResolver)
	orderModule := order.NewModule(gormDB, orderGateway, profileUpdater, orderEventPublisher, idGen, orderContactResolver, orderMerchantID, cfg.InternalAPIToken, logger)

	// Refund module
	orderRepo := orderRepo.NewOrderRepository(gormDB)
	refundCache := cache.NewMemoryRefundCache()
	refundEnquirySLA := cfg.Lazypay.RefundEnquirySLA
	if refundEnquirySLA == 0 {
		refundEnquirySLA = time.Hour
	}
	refundModule := refund.NewModule(
		gormDB, orderRepo, orderGateway, refundGateway, refundCache, profileUpdater,
		refundEventPublisher, mc, logger, idGen, refundEnquirySLA,
	)

	// ═══════════════════════════════════════
	// 8.5. Kafka consumers (when enabled)
	// ═══════════════════════════════════════
	var orderReader, refundReader *kafka.TopicReader
	var orderDLQWriter, refundDLQWriter *kafka.TopicWriter
	var consumerWg sync.WaitGroup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Enabled {
		topics := cfg.Kafka.Topics
		groups := cfg.Kafka.ConsumerGroups
		consCfg := cfg.Kafka.Consumer
		prodCfg := cfg.Kafka.Producer

		orderReader = kafka.NewTopicReader(cfg.Kafka.Brokers, topics.OrderCallback, groups.OrderCallback, consCfg, logger)
		refundReader = kafka.NewTopicReader(cfg.Kafka.Brokers, topics.RefundCallback, groups.RefundCallback, consCfg, logger)
		orderDLQWriter = kafka.NewTopicWriter(cfg.Kafka.Brokers, topics.OrderCallbackDLQ, prodCfg, logger)
		refundDLQWriter = kafka.NewTopicWriter(cfg.Kafka.Brokers, topics.RefundCallbackDLQ, prodCfg, logger)

		orderConsumer := kafkaConsumer.NewOrderCallbackConsumer(
			orderModule.Service, orderDLQWriter, mc, logger, consCfg.MaxRetries)
		refundConsumer := kafkaConsumer.NewRefundCallbackConsumer(
			refundModule.Service, refundDLQWriter, mc, logger, consCfg.MaxRetries)

		consumerWg.Add(1)
		go func() {
			defer consumerWg.Done()
			defer func() {
				if r := recover(); r != nil {
					logger.Error("panic in order callback consumer")
				}
			}()
			if err := orderReader.Consume(ctx, orderConsumer.Handle); err != nil && ctx.Err() == nil {
				logger.Error("order callback consumer stopped", baseLogger.ErrorCode(err.Error()))
			}
		}()

		consumerWg.Add(1)
		go func() {
			defer consumerWg.Done()
			defer func() {
				if r := recover(); r != nil {
					logger.Error("panic in refund callback consumer")
				}
			}()
			if err := refundReader.Consume(ctx, refundConsumer.Handle); err != nil && ctx.Err() == nil {
				logger.Error("refund callback consumer stopped", baseLogger.ErrorCode(err.Error()))
			}
		}()
		logger.Info("Kafka consumers started")
	}

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

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP shutdown error", baseLogger.ErrorCode(err.Error()))
	}
	logger.Info("HTTP server stopped")

	if orderReader != nil {
		orderReader.Close()
	}
	if refundReader != nil {
		refundReader.Close()
	}
	consumerWg.Wait()
	logger.Info("Kafka consumers drained")

	if orderDLQWriter != nil {
		orderDLQWriter.Close()
	}
	if refundDLQWriter != nil {
		refundDLQWriter.Close()
	}

	logger.Info("shutdown complete")
}
