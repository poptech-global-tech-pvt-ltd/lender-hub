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
)

// Module wires together all order module components
type Module struct {
	Service *service.OrderService
}

// NewModule creates a new order module with dependencies
func NewModule(
	db *gorm.DB,
	gw port.OrderGateway,
	profileUpdater *profileService.ProfileUpdater,
	publisher port.OrderEventPublisher,
	idgen *idgen.Generator,
	contactResolver *profileService.UserContactResolver,
) *Module {
	orderRepo := repository.NewOrderRepository(db)
	mappingRepo := repository.NewPaymentMappingRepository(db)
	idempotencyRepo := repository.NewIdempotencyRepository(db)
	idempotencySvc := service.NewIdempotencyService(idempotencyRepo)
	svc := service.NewOrderService(orderRepo, mappingRepo, idempotencySvc, gw, profileUpdater, publisher, idgen, contactResolver)
	return &Module{
		Service: svc,
	}
}

// NewModuleWithStubs creates a new order module with stub implementations
func NewModuleWithStubs(db *gorm.DB, profileUpdater *profileService.ProfileUpdater, idgen *idgen.Generator, contactResolver *profileService.UserContactResolver) *Module {
	gw := stub.NewStubOrderGateway()
	publisher := stub.NewStubOrderEventPublisher()
	return NewModule(db, gw, profileUpdater, publisher, idgen, contactResolver)
}

// RegisterRoutes registers order module routes
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	createHandler := handler.NewCreateOrderHandler(m.Service)
	getHandler := handler.NewGetOrderHandler(m.Service)
	callbackHandler := handler.NewOrderCallbackHandler(m.Service)

	rg.POST("/order", createHandler.Handle)
	rg.GET("/order/:paymentId", getHandler.Handle)
	rg.POST("/callback/order", callbackHandler.Handle)
}
