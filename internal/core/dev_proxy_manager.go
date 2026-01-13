package core

import (
	"crypto/sha256"
	"embed"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"

	"dx/internal/core/domain"
	"dx/internal/ports"

	"gopkg.in/yaml.v3"
)

//go:embed templates/dev-proxy/*/*.tpl
var templateFiles embed.FS

type DevProxyManager struct {
	configRepository         ConfigRepository
	fileService              ports.FileSystem
	containerImageRepository ports.ContainerImageRepository
	containerOrchestrator    ports.ContainerOrchestrator
}

func ProvideDevProxyManager(
	configRepository ConfigRepository,
	fileService ports.FileSystem,
	containerImageRepository ports.ContainerImageRepository,
	containerOrchestrator ports.ContainerOrchestrator,
) *DevProxyManager {
	return &DevProxyManager{
		configRepository:         configRepository,
		fileService:              fileService,
		containerImageRepository: containerImageRepository,
		containerOrchestrator:    containerOrchestrator,
	}
}

// SaveConfiguration saves the given configuration to the HAProxy configuration file to $HOME/.dx/$CONTEXT_NAME/haproxy/haproxy.cfg
func (d *DevProxyManager) SaveConfiguration() error {
	configContext, err := d.configRepository.LoadCurrentConfigurationContext()
	if err != nil {
		return err
	}
	frontendPort := 8080
	proxyPort := 18080
	services := make([]map[string]interface{}, len(configContext.LocalServices))
	for i, localService := range configContext.LocalServices {
		services[i] = map[string]interface{}{
			"Name":            localService.Name,
			"FrontendPort":    frontendPort,
			"ProxyPort":       proxyPort,
			"KubernetesPort":  localService.KubernetesPort,
			"LocalPort":       localService.LocalPort,
			"HealthCheckPath": localService.HealthCheckPath,
			"Selector":        localService.Selector,
		}
		frontendPort++
		proxyPort++
	}

	hash := sha256.New()
	serviceJSON, _ := json.Marshal(configContext.LocalServices)
	hash.Write(serviceJSON)
	checksum := fmt.Sprintf("%x", hash.Sum(nil))[:62]

	values := map[string]interface{}{
		"Services": services,
		"Name":     configContext.Name,
		"Checksum": checksum,
	}

	haproxyConfig, err := renderTemplate("templates/dev-proxy/haproxy/haproxy.cfg.tpl", values)
	if err != nil {
		return err
	}
	err = d.fileService.WriteFile(
		filepath.Join("~", ".dx", configContext.Name, "dev-proxy", "haproxy", "haproxy.cfg"),
		haproxyConfig,
		ports.ReadAllWriteOwner,
	)
	if err != nil {
		return err
	}

	haProxyDockerFile, err := renderTemplate(
		"templates/dev-proxy/haproxy/Dockerfile.tpl",
		values,
	)
	if err != nil {
		return err
	}
	err = d.fileService.WriteFile(
		filepath.Join("~", ".dx", configContext.Name, "dev-proxy", "haproxy", "Dockerfile"),
		haProxyDockerFile,
		ports.ReadWrite,
	)
	if err != nil {
		return err
	}

	mitmProxyDockerFile, err := renderTemplate(
		"templates/dev-proxy/mitmproxy/Dockerfile.tpl",
		values,
	)
	if err != nil {
		return err
	}
	err = d.fileService.WriteFile(
		filepath.Join("~", ".dx", configContext.Name, "dev-proxy", "mitmproxy", "Dockerfile"),
		mitmProxyDockerFile,
		ports.ReadWrite,
	)
	if err != nil {
		return err
	}

	helmManifest, err := renderTemplate("templates/dev-proxy/helm/Chart.yaml.tpl", values)
	if err != nil {
		return err
	}
	err = d.fileService.WriteFile(
		filepath.Join("~", ".dx", configContext.Name, "dev-proxy", "helm", "Chart.yaml"),
		helmManifest,
		ports.ReadWrite,
	)
	if err != nil {
		return err
	}

	devProxyManifests, err := renderTemplate(
		"templates/dev-proxy/helm/deployment.yaml.tpl",
		values,
	)
	if err != nil {
		return err
	}
	err = d.fileService.WriteFile(
		filepath.Join("~", ".dx", configContext.Name, "dev-proxy", "helm", "templates", "dev-proxy.yaml"),
		devProxyManifests,
		ports.ReadWrite,
	)
	if err != nil {
		return err
	}

	return nil
}

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

var templateFunctions = template.FuncMap{
	"toYaml": func(v interface{}) string {
		var buf strings.Builder
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)
		if err := encoder.Encode(v); err != nil {
			return ""
		}
		return buf.String()
	},
	"indent": func(indent int, s string) string {
		lines := strings.Split(s, "\n")
		for i, line := range lines {
			lines[i] = strings.Repeat(" ", indent) + line
		}
		return strings.Join(lines, "\n")
	},
}

func renderTemplate(templatePath string, values map[string]interface{}) ([]byte, error) {
	templateFile, err := templateFiles.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	tmpl, err := template.New(templatePath).Funcs(templateFunctions).Parse(string(templateFile))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, values); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return []byte(result.String()), nil
}
