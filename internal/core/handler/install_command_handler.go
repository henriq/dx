package handler

import (
	"fmt"
	"slices"

	"dx/internal/core"
	"dx/internal/ports"
)

type InstallCommandHandler struct {
	configRepository         core.ConfigRepository
	containerImageRepository ports.ContainerImageRepository
	containerOrchestrator    ports.ContainerOrchestrator
	devProxyManager          *core.DevProxyManager
	environmentEnsurer       core.EnvironmentEnsurer
	scm                      ports.Scm
}

func ProvideInstallCommandHandler(
	configRepository core.ConfigRepository,
	containerImageRepository ports.ContainerImageRepository,
	containerOrchestrator ports.ContainerOrchestrator,
	devProxyManager *core.DevProxyManager,
	environmentEnsurer core.EnvironmentEnsurer,
	scm ports.Scm,
) InstallCommandHandler {
	return InstallCommandHandler{
		configRepository:         configRepository,
		containerImageRepository: containerImageRepository,
		containerOrchestrator:    containerOrchestrator,
		devProxyManager:          devProxyManager,
		environmentEnsurer:       environmentEnsurer,
		scm:                      scm,
	}
}

func (h *InstallCommandHandler) Handle(services []string, selectedProfile string, skipDevProxy bool) error {
	err := h.environmentEnsurer.EnsureExpectedClusterIsSelected()
	if err != nil {
		return err
	}

	if !skipDevProxy {
		err = h.devProxyManager.SaveConfiguration()
		if err != nil {
			return err
		}

		err = h.devProxyManager.BuildDevProxy()
		if err != nil {
			return err
		}

		err = h.devProxyManager.InstallDevProxy()
		if err != nil {
			return err
		}
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

		err := h.scm.Download(service.HelmRepoPath, service.HelmBranch, service.HelmPath)
		if err != nil {
			return err
		}
		if err = h.containerOrchestrator.InstallService(&service); err != nil {
			return fmt.Errorf("failed to install service %s: %v", service.Name, err)
		}
	}

	return nil
}
