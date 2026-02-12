package profile

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"lending-hub-service/internal/domain/profile/handler"
	"lending-hub-service/internal/domain/profile/port"
	"lending-hub-service/internal/domain/profile/repository"
	"lending-hub-service/internal/domain/profile/service"
	"lending-hub-service/internal/domain/profile/stub"
)

// Module wires together all profile module components
type Module struct {
	Service service.ProfileService
	Updater *service.ProfileUpdater
}

// NewModule creates a new profile module with dependencies
func NewModule(db *gorm.DB, gw port.ProfileGateway, cache port.ProfileCache, publisher port.ProfileEventPublisher) *Module {
	repo := repository.NewProfileRepository(db)
	svc := service.NewProfileService(repo, gw, cache)
	updater := service.NewProfileUpdater(repo, publisher)
	return &Module{
		Service: svc,
		Updater: updater,
	}
}

// NewModuleWithStubs creates a new profile module with stub implementations
func NewModuleWithStubs(db *gorm.DB) *Module {
	gw := stub.NewStubProfileGateway()
	cache := stub.NewStubProfileCache()
	publisher := stub.NewStubProfileEventPublisher()
	return NewModule(db, gw, cache, publisher)
}

// RegisterRoutes registers profile module routes
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	h := handler.NewCustomerStatusHandler(m.Service)
	rg.POST("/customer-status", h.Handle)
}
