package refund

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	orderPort "lending-hub-service/internal/domain/order/port"
	"lending-hub-service/internal/domain/refund/handler"
	refundPort "lending-hub-service/internal/domain/refund/port"
	"lending-hub-service/internal/domain/refund/repository"
	"lending-hub-service/internal/domain/refund/service"
	"lending-hub-service/internal/infrastructure/observability/metrics"
	profileService "lending-hub-service/internal/domain/profile/service"
	baseLogger "lending-hub-service/pkg/logger"
	"lending-hub-service/pkg/idgen"
)

// Module wires together all refund module components
type Module struct {
	createHandler        *handler.CreateRefundHandler
	getHandler           *handler.GetRefundHandler
	getByRefundIDHandler *handler.GetRefundByRefundIDHandler
	listForOrderHandler  *handler.ListRefundsForOrderHandler
	listByUserHandler    *handler.ListRefundsByUserHandler
}

// NewModule creates a new refund module with dependencies
func NewModule(
	db *gorm.DB,
	orderRepo orderPort.OrderRepository,
	orderGateway orderPort.OrderGateway,
	gateway refundPort.RefundGateway,
	cache refundPort.RefundCache,
	profileUpdater *profileService.ProfileUpdater,
	mc metrics.MetricsClient,
	logger *baseLogger.Logger,
	idgen *idgen.Generator,
	enquirySLA time.Duration,
) *Module {
	repo := repository.NewRefundRepository(db)
	enquirySvc := service.NewRefundEnquiryService(gateway, repo, cache, mc, logger, enquirySLA)
	refundSvc := service.NewRefundService(repo, orderRepo, orderGateway, gateway, cache, enquirySvc, profileUpdater, mc, logger, idgen)

	return &Module{
		createHandler:        handler.NewCreateRefundHandler(refundSvc),
		getHandler:           handler.NewGetRefundHandler(refundSvc),
		getByRefundIDHandler: handler.NewGetRefundByRefundIDHandler(refundSvc),
		listForOrderHandler:  handler.NewListRefundsForOrderHandler(refundSvc),
		listByUserHandler:    handler.NewListRefundsByUserHandler(refundSvc),
	}
}

// RegisterRoutes registers refund routes (static segments before parameterised)
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/refund", m.createHandler.Handle)
	rg.GET("/refund/loan/:refundId", m.getByRefundIDHandler.Handle)
	rg.GET("/refund/:paymentRefundId", m.getHandler.Handle)
	rg.GET("/refunds", m.listForOrderHandler.Handle)
	rg.GET("/refunds/user", m.listByUserHandler.Handle)
}
