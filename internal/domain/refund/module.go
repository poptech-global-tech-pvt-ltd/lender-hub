package refund

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	orderPort "lending-hub-service/internal/domain/order/port"
	orderRepo "lending-hub-service/internal/domain/order/repository"
	"lending-hub-service/internal/domain/refund/handler"
	refundPort "lending-hub-service/internal/domain/refund/port"
	"lending-hub-service/internal/domain/refund/repository"
	"lending-hub-service/internal/domain/refund/service"
	"lending-hub-service/internal/domain/refund/stub"
	profileService "lending-hub-service/internal/domain/profile/service"
)

// Module wires together all refund module components
type Module struct {
	Service *service.RefundService
}

// NewModule creates a new refund module with dependencies
func NewModule(
	db *gorm.DB,
	gw refundPort.RefundGateway,
	orderRepo orderPort.OrderRepository,
	profileUpdater *profileService.ProfileUpdater,
) *Module {
	refundRepo := repository.NewRefundRepository(db)
	svc := service.NewRefundService(refundRepo, orderRepo, gw, profileUpdater)
	return &Module{
		Service: svc,
	}
}

// NewModuleWithStubs creates a new refund module with stub implementations
func NewModuleWithStubs(db *gorm.DB, profileUpdater *profileService.ProfileUpdater) *Module {
	gw := stub.NewStubRefundGateway()
	orderRepo := orderRepo.NewOrderRepository(db)
	return NewModule(db, gw, orderRepo, profileUpdater)
}

// RegisterRoutes registers refund module routes
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	createHandler := handler.NewCreateRefundHandler(m.Service)
	callbackHandler := handler.NewRefundCallbackHandler(m.Service)
	getHandler := handler.NewGetRefundHandler(m.Service)

	rg.POST("/refund", createHandler.Handle)
	rg.GET("/refund/:refundId", getHandler.Handle)
	rg.POST("/callback/refund", callbackHandler.Handle)
}
