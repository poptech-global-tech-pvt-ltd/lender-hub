package profile

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lending-hub-service/internal/domain/profile/handler"
	"lending-hub-service/internal/domain/profile/port"
	"lending-hub-service/internal/domain/profile/repository"
	"lending-hub-service/internal/domain/profile/service"
	"lending-hub-service/internal/domain/profile/stub"
	"lending-hub-service/internal/infrastructure/userprofile"
	baseLogger "lending-hub-service/pkg/logger"
)

// Module wires together all profile module components
type Module struct {
	Service                service.ProfileService
	Updater                *service.ProfileUpdater
	eligibilityHandler     *handler.EligibilityHandler
	customerStatusHandler  *handler.CustomerStatusHandler
	combinedProfileHandler *handler.CombinedProfileHandler
}

// NewModule creates a new profile module with dependencies
func NewModule(
	db *gorm.DB,
	gw port.ProfileGateway,
	publisher port.ProfileEventPublisher,
	contactResolver *service.UserContactResolver,
	profileClient *userprofile.Client,
	logger *baseLogger.Logger,
) *Module {
	repo := repository.NewProfileRepository(db)
	svc := service.NewProfileService(gw, contactResolver, repo, profileClient, logger)
	updater := service.NewProfileUpdater(repo, publisher)

	return &Module{
		Service:                svc,
		Updater:                updater,
		eligibilityHandler:     handler.NewEligibilityHandler(svc),
		customerStatusHandler:  handler.NewCustomerStatusHandler(svc),
		combinedProfileHandler: handler.NewCombinedProfileHandler(svc),
	}
}

// NewModuleWithStubs creates a new profile module with stub implementations
func NewModuleWithStubs(db *gorm.DB, contactResolver *service.UserContactResolver, profileClient *userprofile.Client, logger *baseLogger.Logger) *Module {
	gw := stub.NewStubProfileGateway()
	publisher := stub.NewStubProfileEventPublisher()
	return NewModule(db, gw, publisher, contactResolver, profileClient, logger)
}

// RegisterRoutes registers profile module routes
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/eligibility", m.eligibilityHandler.Handle)
	rg.POST("/customer-status", m.customerStatusHandler.Handle)
	rg.GET("/profile/:userId", m.combinedProfileHandler.Handle)
}
