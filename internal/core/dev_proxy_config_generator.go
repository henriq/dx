package core

import (
	"crypto/sha256"
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"dx/internal/core/domain"

	"gopkg.in/yaml.v3"
)

const (
	// devProxyFrontendStartPort is the starting port for HAProxy frontend listeners.
	devProxyFrontendStartPort = 8080
	// devProxyProxyStartPort is the starting port for mitmproxy backends.
	devProxyProxyStartPort = 18080
)

//go:embed templates/dev-proxy/*/*.tpl
var templateFiles embed.FS

// DevProxyConfigs holds all generated configuration files for the dev-proxy.
type DevProxyConfigs struct {
	HAProxyConfig       []byte
	HAProxyDockerfile   []byte
	MitmProxyDockerfile []byte
	HelmChartYaml       []byte
	HelmDeploymentYaml  []byte
}

// DevProxyConfigGenerator generates dev-proxy configuration files from domain configuration.
// This is pure business logic with no I/O operations.
type DevProxyConfigGenerator struct{}

// ProvideDevProxyConfigGenerator creates a new DevProxyConfigGenerator.
func ProvideDevProxyConfigGenerator() *DevProxyConfigGenerator {
	return &DevProxyConfigGenerator{}
}

// Generate creates all dev-proxy configuration files from the given configuration context.
// Returns a DevProxyConfigs struct containing all generated content.
func (g *DevProxyConfigGenerator) Generate(configContext *domain.ConfigurationContext) (*DevProxyConfigs, error) {
	values := g.buildTemplateValues(configContext)

	haproxyConfig, err := renderTemplate("templates/dev-proxy/haproxy/haproxy.cfg.tpl", values)
	if err != nil {
		return nil, fmt.Errorf("failed to render haproxy config: %w", err)
	}

	haproxyDockerfile, err := renderTemplate("templates/dev-proxy/haproxy/Dockerfile.tpl", values)
	if err != nil {
		return nil, fmt.Errorf("failed to render haproxy dockerfile: %w", err)
	}

	mitmproxyDockerfile, err := renderTemplate("templates/dev-proxy/mitmproxy/Dockerfile.tpl", values)
	if err != nil {
		return nil, fmt.Errorf("failed to render mitmproxy dockerfile: %w", err)
	}

	helmChartYaml, err := renderTemplate("templates/dev-proxy/helm/Chart.yaml.tpl", values)
	if err != nil {
		return nil, fmt.Errorf("failed to render helm chart.yaml: %w", err)
	}

	helmDeploymentYaml, err := renderTemplate("templates/dev-proxy/helm/deployment.yaml.tpl", values)
	if err != nil {
		return nil, fmt.Errorf("failed to render helm deployment.yaml: %w", err)
	}

	return &DevProxyConfigs{
		HAProxyConfig:       haproxyConfig,
		HAProxyDockerfile:   haproxyDockerfile,
		MitmProxyDockerfile: mitmproxyDockerfile,
		HelmChartYaml:       helmChartYaml,
		HelmDeploymentYaml:  helmDeploymentYaml,
	}, nil
}

// buildTemplateValues constructs the values map for template rendering.
func (g *DevProxyConfigGenerator) buildTemplateValues(configContext *domain.ConfigurationContext) map[string]interface{} {
	frontendPort := devProxyFrontendStartPort
	proxyPort := devProxyProxyStartPort
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
	// Error can be safely ignored: LocalServices contains only JSON-serializable primitive types
	// (strings, ints, and map[string]string). json.Marshal cannot fail for these types.
	serviceJSON, _ := json.Marshal(configContext.LocalServices)
	hash.Write(serviceJSON)
	// Truncate to 62 characters to stay within Kubernetes label value limit (63 chars max).
	// SHA256 produces 64 hex characters; this provides sufficient uniqueness for checksum purposes.
	checksum := fmt.Sprintf("%x", hash.Sum(nil))[:62]

	return map[string]interface{}{
		"Services": services,
		"Name":     configContext.Name,
		"Checksum": checksum,
	}
}

var templateFunctions = template.FuncMap{
	"toYaml": func(v interface{}) string {
		var buf strings.Builder
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(2)
		if err := encoder.Encode(v); err != nil {
			// Return error marker that will be visible in output and cause YAML parsing to fail.
			// This is preferable to silent failure with empty string.
			return fmt.Sprintf("# ERROR: failed to encode YAML: %v", err)
		}
		return buf.String()
	},
	"indent": func(indent int, s string) string {
		// Normalize line endings to handle both Unix (\n) and Windows (\r\n)
		s = strings.ReplaceAll(s, "\r\n", "\n")
		s = strings.ReplaceAll(s, "\r", "\n")
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
