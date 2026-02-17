package onboarding

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lending-hub-service/internal/domain/onboarding/handler"
	"lending-hub-service/internal/domain/onboarding/port"
	"lending-hub-service/internal/domain/onboarding/repository"
	"lending-hub-service/internal/domain/onboarding/service"
	"lending-hub-service/internal/domain/onboarding/stub"
	profileService "lending-hub-service/internal/domain/profile/service"
	"lending-hub-service/pkg/idgen"
)

// Module wires together all onboarding module components
type Module struct {
	Service *service.OnboardingService
}

// NewModule creates a new onboarding module with dependencies
func NewModule(db *gorm.DB, gw port.OnboardingGateway, profileUpdater *profileService.ProfileUpdater, idgen *idgen.Generator) *Module {
	repo := repository.NewOnboardingRepository(db)
	eventStore := repository.NewOnboardingEventStore(db)
	svc := service.NewOnboardingService(repo, eventStore, gw, profileUpdater, idgen)
	return &Module{
		Service: svc,
	}
}

// NewModuleWithStubs creates a new onboarding module with stub implementations
func NewModuleWithStubs(db *gorm.DB, profileUpdater *profileService.ProfileUpdater, idgen *idgen.Generator) *Module {
	gw := stub.NewStubOnboardingGateway()
	return NewModule(db, gw, profileUpdater, idgen)
}

// RegisterRoutes registers onboarding module routes
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	startHandler := handler.NewStartOnboardingHandler(m.Service)
	statusHandler := handler.NewOnboardingStatusHandler(m.Service)
	callbackHandler := handler.NewOnboardingCallbackHandler(m.Service)

	rg.POST("/onboarding", startHandler.Handle)
	rg.GET("/onboarding/status", statusHandler.Handle)
	rg.POST("/callback/onboarding", callbackHandler.Handle)
}
