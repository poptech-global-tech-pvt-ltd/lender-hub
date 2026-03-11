package order

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lending-hub-service/internal/domain/order/handler"
	"lending-hub-service/internal/domain/order/port"
	"lending-hub-service/internal/domain/order/repository"
	"lending-hub-service/internal/domain/order/service"
	"lending-hub-service/internal/domain/order/stub"
	refundPort "lending-hub-service/internal/domain/refund/port"
	"lending-hub-service/pkg/idgen"
	baseLogger "lending-hub-service/pkg/logger"
)

// Module wires together all order module components
type Module struct {
	Service          *service.OrderService
	summaryHandler   *handler.OrderSummaryHandler
	internalAPIToken string
	logger           *baseLogger.Logger
}

// NewModule creates a new order module with dependencies
func NewModule(
	db *gorm.DB,
	gw port.OrderGateway,
	profileUpdater port.ProfileUpdater,
	publisher port.OrderEventPublisher,
	idgen *idgen.Generator,
	contactResolver port.ContactResolver,
	merchantID string,
	internalAPIToken string,
	refundRepo refundPort.RefundRepository,
	logger *baseLogger.Logger,
) *Module {
	orderRepo := repository.NewOrderRepository(db)
	mappingRepo := repository.NewPaymentMappingRepository(db)
	idempotencyRepo := repository.NewIdempotencyRepository(db)
	idempotencySvc := service.NewIdempotencyService(idempotencyRepo)
	svc := service.NewOrderService(orderRepo, mappingRepo, idempotencySvc, gw, profileUpdater, publisher, idgen, contactResolver, merchantID, logger)
	summarySvc := service.NewOrderSummaryService(svc, refundRepo, logger)
	return &Module{
		Service:          svc,
		summaryHandler:   handler.NewOrderSummaryHandler(summarySvc),
		internalAPIToken: internalAPIToken,
		logger:           logger,
	}
}

// NewStubOrderEventPublisher returns a no-op event publisher for use when Kafka is disabled
func NewStubOrderEventPublisher() port.OrderEventPublisher {
	return stub.NewStubOrderEventPublisher()
}

// NewModuleWithStubs creates a new order module with stub implementations
func NewModuleWithStubs(db *gorm.DB, profileUpdater port.ProfileUpdater, idgen *idgen.Generator, contactResolver port.ContactResolver, merchantID, internalAPIToken string, refundRepo refundPort.RefundRepository, logger *baseLogger.Logger) *Module {
	gw := stub.NewStubOrderGateway()
	publisher := stub.NewStubOrderEventPublisher()
	return NewModule(db, gw, profileUpdater, publisher, idgen, contactResolver, merchantID, internalAPIToken, refundRepo, logger)
}

// RegisterRoutes registers order module routes
// NOTE: No POST /callback/order — callbacks come via Kafka consumer
// Route order: more specific paths first (loan, recon, summary) before /order/:paymentId
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	createHandler := handler.NewCreateOrderHandler(m.Service, m.logger)
	getHandler := handler.NewGetOrderHandler(m.Service)
	getByLoanHandler := handler.NewGetOrderByLoanHandler(m.Service)
	getByReconHandler := handler.NewGetOrderByReconHandler(m.Service)
	listHandler := handler.NewListOrdersHandler(m.Service)
	supportHandler := handler.NewSupportOrderHandler(m.Service, m.internalAPIToken)
	supportByLoanHandler := handler.NewSupportOrderByLoanHandler(m.Service, m.internalAPIToken)

	rg.POST("/order", createHandler.Handle)
	rg.GET("/order/loan/:loanId", getByLoanHandler.Handle)
	rg.GET("/order/recon/:lenderOrderId", getByReconHandler.Handle)
	rg.GET("/order/:paymentId/summary", m.summaryHandler.Handle)
	rg.GET("/order/:paymentId", getHandler.Handle)
	rg.GET("/orders", listHandler.Handle)
	rg.PATCH("/order/loan/:loanId/status", supportByLoanHandler.Handle)
	rg.PATCH("/order/:paymentId/status", supportHandler.Handle)
}
