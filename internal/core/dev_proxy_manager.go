package core

import (
	"fmt"
	"path/filepath"

	"dx/internal/core/domain"
	"dx/internal/ports"
)

// DevProxyManager orchestrates dev-proxy operations including configuration saving,
// image building, and service installation/uninstallation.
type DevProxyManager struct {
	configRepository         ConfigRepository
	fileService              ports.FileSystem
	containerImageRepository ports.ContainerImageRepository
	containerOrchestrator    ports.ContainerOrchestrator
	configGenerator          *DevProxyConfigGenerator
}

// ProvideDevProxyManager creates a new DevProxyManager with all required dependencies.
func ProvideDevProxyManager(
	configRepository ConfigRepository,
	fileService ports.FileSystem,
	containerImageRepository ports.ContainerImageRepository,
	containerOrchestrator ports.ContainerOrchestrator,
	configGenerator *DevProxyConfigGenerator,
) *DevProxyManager {
	return &DevProxyManager{
		configRepository:         configRepository,
		fileService:              fileService,
		containerImageRepository: containerImageRepository,
		containerOrchestrator:    containerOrchestrator,
		configGenerator:          configGenerator,
	}
}

// SaveConfiguration generates and saves all dev-proxy configuration files
// to $HOME/.dx/$CONTEXT_NAME/dev-proxy/
func (d *DevProxyManager) SaveConfiguration() error {
	configContext, err := d.configRepository.LoadCurrentConfigurationContext()
	if err != nil {
		return err
	}

	configs, err := d.configGenerator.Generate(configContext)
	if err != nil {
		return err
	}

	basePath := filepath.Join("~", ".dx", configContext.Name, "dev-proxy")

	// Write HAProxy config
	err = d.fileService.WriteFile(
		filepath.Join(basePath, "haproxy", "haproxy.cfg"),
		configs.HAProxyConfig,
		ports.ReadAllWriteOwner,
	)
	if err != nil {
		return err
	}

	// Write HAProxy Dockerfile
	err = d.fileService.WriteFile(
		filepath.Join(basePath, "haproxy", "Dockerfile"),
		configs.HAProxyDockerfile,
		ports.ReadWrite,
	)
	if err != nil {
		return err
	}

	// Write mitmproxy Dockerfile
	err = d.fileService.WriteFile(
		filepath.Join(basePath, "mitmproxy", "Dockerfile"),
		configs.MitmProxyDockerfile,
		ports.ReadWrite,
	)
	if err != nil {
		return err
	}

	// Write Helm Chart.yaml
	err = d.fileService.WriteFile(
		filepath.Join(basePath, "helm", "Chart.yaml"),
		configs.HelmChartYaml,
		ports.ReadWrite,
	)
	if err != nil {
		return err
	}

	// Write Helm deployment manifest
	err = d.fileService.WriteFile(
		filepath.Join(basePath, "helm", "templates", "dev-proxy.yaml"),
		configs.HelmDeploymentYaml,
		ports.ReadWrite,
	)
	if err != nil {
		return err
	}

	return nil
}

// BuildDevProxy builds the HAProxy and mitmproxy Docker images for the dev-proxy.
func (d *DevProxyManager) BuildDevProxy() error {
	configContext, err := d.configRepository.LoadCurrentConfigurationContext()
	if err != nil {
		return err
	}
	homeDir, err := d.fileService.HomeDir()
	if err != nil {
		return err
	}
	dockerImages := []domain.DockerImage{
		{
			Name:                     fmt.Sprintf("henriq/haproxy-%s", configContext.Name),
			DockerfilePath:           "Dockerfile",
			BuildContextRelativePath: ".",
			Path:                     filepath.Join(homeDir, ".dx", configContext.Name, "dev-proxy", "haproxy"),
		},
		{
			Name:                     fmt.Sprintf("henriq/mitmproxy-%s", configContext.Name),
			DockerfilePath:           "Dockerfile",
			BuildContextRelativePath: ".",
			Path:                     filepath.Join(homeDir, ".dx", configContext.Name, "dev-proxy", "mitmproxy"),
		},
	}

	for _, image := range dockerImages {
		err = d.containerImageRepository.BuildImage(image)
		if err != nil {
			return fmt.Errorf("failed to build image %s: %w", image.Name, err)
		}
	}

	return nil
}

// InstallDevProxy installs the dev-proxy service to Kubernetes using Helm.
func (d *DevProxyManager) InstallDevProxy() error {
	configContext, err := d.configRepository.LoadCurrentConfigurationContext()
	if err != nil {
		return err
	}
	homeDir, err := d.fileService.HomeDir()
	if err != nil {
		return err
	}

	service := domain.Service{
		Name:     "dev-proxy",
		HelmPath: filepath.Join(homeDir, ".dx", configContext.Name, "dev-proxy", "helm"),
	}
	return d.containerOrchestrator.InstallDevProxy(&service)
}

// UninstallDevProxy removes the dev-proxy service from Kubernetes.
func (d *DevProxyManager) UninstallDevProxy() error {
	configContext, err := d.configRepository.LoadCurrentConfigurationContext()
	if err != nil {
		return err
	}
	homeDir, err := d.fileService.HomeDir()
	if err != nil {
		return err
	}

	service := domain.Service{
		Name:     "dev-proxy",
		HelmPath: filepath.Join(homeDir, ".dx", configContext.Name, "dev-proxy", "helm"),
	}
	return d.containerOrchestrator.UninstallService(&service)
}
