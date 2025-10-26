package ports

import (
	"dx/internal/core/domain"
)

type ContainerOrchestrator interface {
	CreateClusterEnvironmentKey() (string, error)
	InstallService(service *domain.Service) error
	InstallDevProxy(service *domain.Service) error
	UninstallService(service *domain.Service) error
	HasDeployedServices() (bool, error)
}
