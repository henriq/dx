package handler

import (
	"slices"

	"dx/internal/core"
	"dx/internal/ports"
)

type UninstallCommandHandler struct {
	configRepository      core.ConfigRepository
	containerOrchestrator ports.ContainerOrchestrator
	environmentEnsurer    core.EnvironmentEnsurer
	devProxyManager       *core.DevProxyManager
}

func ProvideUninstallCommandHandler(
	configRepository core.ConfigRepository,
	containerOrchestrator ports.ContainerOrchestrator,
	environmentEnsurer core.EnvironmentEnsurer,
	devProxyManager *core.DevProxyManager,
) UninstallCommandHandler {
	return UninstallCommandHandler{
		configRepository:      configRepository,
		containerOrchestrator: containerOrchestrator,
		environmentEnsurer:    environmentEnsurer,
		devProxyManager:       devProxyManager,
	}
}

func (h *UninstallCommandHandler) Handle(services []string, selectedProfile string) error {
	err := h.environmentEnsurer.EnsureExpectedClusterIsSelected()
	if err != nil {
		return err
	}

	configContext, err := h.configRepository.LoadCurrentConfigurationContext()
	if err != nil {
		return err
	}

	for _, service := range configContext.Services {
		if len(services) == 0 && !slices.Contains(service.Profiles, selectedProfile) {
			continue
		}

		if len(services) > 0 && !slices.ContainsFunc(services, func(s string) bool { return s == service.Name }) {
			continue
		}

		_ = h.containerOrchestrator.UninstallService(&service)
	}

	hasDeployedServices, err := h.containerOrchestrator.HasDeployedServices()
	if err != nil {
		return err
	}
	if !hasDeployedServices {
		err = h.devProxyManager.UninstallDevProxy()
		if err != nil {
			return err
		}
	}

	return nil
}
