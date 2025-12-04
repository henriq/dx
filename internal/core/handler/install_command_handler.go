package handler

import (
	"fmt"
	"slices"

	"dx/internal/cli/output"
	"dx/internal/cli/progress"
	"dx/internal/core"
	"dx/internal/core/domain"
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
		output.PrintHeader("Setting up dev-proxy")
		fmt.Println()

		proxySteps := []string{"Save configuration", "Build dev-proxy", "Install dev-proxy"}
		tracker := progress.NewTrackerWithVerb(proxySteps, "Running")
		tracker.Start()

		tracker.StartItem(0)
		err = h.devProxyManager.SaveConfiguration()
		if err != nil {
			tracker.CompleteItem(0, err)
			tracker.PrintItemComplete(0)
			tracker.Stop()
			return err
		}
		tracker.CompleteItem(0, nil)
		tracker.PrintItemComplete(0)

		tracker.StartItem(1)
		err = h.devProxyManager.BuildDevProxy()
		if err != nil {
			tracker.CompleteItem(1, err)
			tracker.PrintItemComplete(1)
			tracker.Stop()
			return err
		}
		tracker.CompleteItem(1, nil)
		tracker.PrintItemComplete(1)

		tracker.StartItem(2)
		err = h.devProxyManager.InstallDevProxy()
		if err != nil {
			tracker.CompleteItem(2, err)
			tracker.PrintItemComplete(2)
			tracker.Stop()
			return err
		}
		tracker.CompleteItem(2, nil)
		tracker.PrintItemComplete(2)

		tracker.Stop()
		fmt.Println()
	}

	configContext, err := h.configRepository.LoadCurrentConfigurationContext()
	if err != nil {
		return err
	}

	// Collect services to install
	var servicesToInstall []domain.Service
	for _, service := range configContext.Services {
		if len(services) == 0 && !slices.Contains(service.Profiles, selectedProfile) {
			continue
		}

		if len(services) > 0 && !slices.ContainsFunc(services, func(s string) bool { return s == service.Name }) {
			continue
		}

		servicesToInstall = append(servicesToInstall, service)
	}

	if len(servicesToInstall) > 0 {
		output.PrintHeader("Installing services")
		fmt.Println()

		serviceNames := make([]string, len(servicesToInstall))
		for i, svc := range servicesToInstall {
			serviceNames[i] = svc.Name
		}
		tracker := progress.NewTrackerWithVerb(serviceNames, "Installing")
		tracker.Start()

		for i, service := range servicesToInstall {
			tracker.StartItem(i)

			err := h.scm.Download(service.HelmRepoPath, service.HelmBranch, service.HelmPath)
			if err != nil {
				tracker.CompleteItem(i, err)
				tracker.PrintItemComplete(i)
				tracker.Stop()
				return err
			}

			if err = h.containerOrchestrator.InstallService(&service); err != nil {
				installErr := fmt.Errorf("failed to install service %s: %v", service.Name, err)
				tracker.CompleteItem(i, installErr)
				tracker.PrintItemComplete(i)
				tracker.Stop()
				return installErr
			}

			tracker.CompleteItem(i, nil)
			tracker.PrintItemComplete(i)
		}

		tracker.Stop()
		fmt.Println()
		output.PrintSuccess(fmt.Sprintf("Installed %d %s", len(servicesToInstall), output.Plural(len(servicesToInstall), "service", "services")))
	}

	return nil
}
