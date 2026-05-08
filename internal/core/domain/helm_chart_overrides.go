package domain

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var validOverrideServiceName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,62}$`)

// HelmChartOverrides maps a service name to a local chart directory that
// should be rendered in place of the service's configured Helm chart.
type HelmChartOverrides struct {
	chartDirectoryByService map[string]string
}

// ValidateOverrideServiceName returns an error if name is not a permitted
// service-name identifier for a Helm chart override.
func ValidateOverrideServiceName(name string) error {
	if !validOverrideServiceName.MatchString(name) {
		return fmt.Errorf(
			"invalid service name %q; must start with a letter or digit and contain only letters, digits, hyphens, and underscores (max 63 characters)",
			name,
		)
	}
	return nil
}

func NewHelmChartOverrides(chartDirectoryByService map[string]string) (HelmChartOverrides, error) {
	copied := make(map[string]string, len(chartDirectoryByService))
	for serviceName, chartDirectory := range chartDirectoryByService {
		if err := ValidateOverrideServiceName(serviceName); err != nil {
			return HelmChartOverrides{}, err
		}
		copied[serviceName] = chartDirectory
	}
	return HelmChartOverrides{chartDirectoryByService: copied}, nil
}

func (o HelmChartOverrides) IsEmpty() bool {
	return len(o.chartDirectoryByService) == 0
}

func (o HelmChartOverrides) LookupChartDirectory(serviceName string) (string, bool) {
	directory, ok := o.chartDirectoryByService[serviceName]
	return directory, ok
}

// FindUnusedOverrides returns the service names with overrides that are not
// present in the given list, sorted lexically.
func (o HelmChartOverrides) FindUnusedOverrides(servicesInScope []Service) []string {
	if o.IsEmpty() {
		return nil
	}
	inScope := map[string]struct{}{}
	for _, service := range servicesInScope {
		inScope[service.Name] = struct{}{}
	}
	var unused []string
	for serviceName := range o.chartDirectoryByService {
		if _, ok := inScope[serviceName]; !ok {
			unused = append(unused, serviceName)
		}
	}
	sort.Strings(unused)
	return unused
}

// ValidateAgainstServices returns an error if any override targets a service
// name not present in the given list.
func (o HelmChartOverrides) ValidateAgainstServices(services []Service) error {
	if o.IsEmpty() {
		return nil
	}

	knownServices := map[string]struct{}{}
	for _, service := range services {
		knownServices[service.Name] = struct{}{}
	}

	var unknownServices []string
	for serviceName := range o.chartDirectoryByService {
		if _, ok := knownServices[serviceName]; !ok {
			unknownServices = append(unknownServices, serviceName)
		}
	}
	if len(unknownServices) == 0 {
		return nil
	}
	sort.Strings(unknownServices)
	return fmt.Errorf(
		"service(s) not found: %s; available services: %s",
		strings.Join(quoteAll(unknownServices), ", "),
		availableServiceNames(services),
	)
}

func quoteAll(values []string) []string {
	quoted := make([]string, len(values))
	for index, value := range values {
		quoted[index] = fmt.Sprintf("%q", value)
	}
	return quoted
}

func availableServiceNames(services []Service) string {
	names := make([]string, 0, len(services))
	for _, service := range services {
		names = append(names, service.Name)
	}
	return strings.Join(names, ", ")
}
