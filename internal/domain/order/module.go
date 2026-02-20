package order

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lending-hub-service/internal/domain/order/handler"
	"lending-hub-service/internal/domain/order/port"
	"lending-hub-service/internal/domain/order/repository"
	"lending-hub-service/internal/domain/order/service"
	"lending-hub-service/internal/domain/order/stub"
	profileService "lending-hub-service/internal/domain/profile/service"
	"lending-hub-service/pkg/idgen"
	baseLogger "lending-hub-service/pkg/logger"
)

// Module wires together all order module components
type Module struct {
	Service          *service.OrderService
	internalAPIToken string
	logger           *baseLogger.Logger
}

// NewModule creates a new order module with dependencies
func NewModule(
	db *gorm.DB,
	gw port.OrderGateway,
	profileUpdater *profileService.ProfileUpdater,
	publisher port.OrderEventPublisher,
	idgen *idgen.Generator,
	contactResolver *profileService.UserContactResolver,
	merchantID string,
	internalAPIToken string,
	logger *baseLogger.Logger,
) *Module {
	orderRepo := repository.NewOrderRepository(db)
	mappingRepo := repository.NewPaymentMappingRepository(db)
	idempotencyRepo := repository.NewIdempotencyRepository(db)
	idempotencySvc := service.NewIdempotencyService(idempotencyRepo)
	svc := service.NewOrderService(orderRepo, mappingRepo, idempotencySvc, gw, profileUpdater, publisher, idgen, contactResolver, merchantID)
	return &Module{
		Service:          svc,
		internalAPIToken: internalAPIToken,
		logger:           logger,
	}
}

// NewModuleWithStubs creates a new order module with stub implementations
func NewModuleWithStubs(db *gorm.DB, profileUpdater *profileService.ProfileUpdater, idgen *idgen.Generator, contactResolver *profileService.UserContactResolver, merchantID, internalAPIToken string, logger *baseLogger.Logger) *Module {
	gw := stub.NewStubOrderGateway()
	publisher := stub.NewStubOrderEventPublisher()
	return NewModule(db, gw, profileUpdater, publisher, idgen, contactResolver, merchantID, internalAPIToken, logger)
}

// RegisterRoutes registers order module routes
// NOTE: No POST /callback/order — callbacks come via Kafka consumer
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	createHandler := handler.NewCreateOrderHandler(m.Service, m.logger)
	getHandler := handler.NewGetOrderHandler(m.Service)
	listHandler := handler.NewListOrdersHandler(m.Service)
	supportHandler := handler.NewSupportOrderHandler(m.Service, m.internalAPIToken)

	rg.POST("/order", createHandler.Handle)
	rg.GET("/order/:paymentId", getHandler.Handle)
	rg.GET("/orders", listHandler.Handle)
	rg.PATCH("/order/:paymentId/status", supportHandler.Handle)
}
